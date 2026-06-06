package service

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// StartEnterpriseUsageStatsTask 启动企业用量统计定时任务（每小时聚合）
func StartEnterpriseUsageStatsTask() {
	ticker := time.NewTicker(1 * time.Hour)
	logger.SysLog("[EnterpriseUsage] stats aggregation cron started (every hour)")
	for range ticker.C {
		if err := aggregateEnterpriseUsage(); err != nil {
			logger.SysError(fmt.Sprintf("[EnterpriseUsage] aggregation failed: %v", err))
		}
	}
}

// aggregateEnterpriseUsage 聚合 TokenTransaction 数据到 EnterpriseUsageStat
func aggregateEnterpriseUsage() error {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	today := now.Format("2006-01-02")

	// 聚合昨天和今天的数据（兜底昨天的未聚合数据）
	for _, date := range []string{yesterday, today} {
		startOfDay := date + " 00:00:00"
		endOfDay := date + " 23:59:59"

		startUnix := parseTimeToUnix(startOfDay)
		endUnix := parseTimeToUnix(endOfDay)

		// 按 org_id + user_id + model_name 聚合
		type AggRow struct {
			OrgId        int
			UserId       int
			ModelName    string
			TokenCount   int64
			RequestCount int64
			CostCents    int64
		}
		var rows []AggRow

		err := model.DB.Raw(`
			SELECT
				COALESCE(u.organization_id, 0) as org_id,
				t.user_id,
				t.model_name,
				COALESCE(SUM(t.prompt_tokens + t.completion_tokens), 0) as token_count,
				COUNT(*) as request_count,
				COALESCE(SUM(t.total_amount * 100), 0) as cost_cents
			FROM token_transactions t
			LEFT JOIN users u ON u.id = t.user_id
			WHERE t.created_time >= ? AND t.created_time <= ?
				AND u.organization_id > 0
			GROUP BY u.organization_id, t.user_id, t.model_name
		`, startUnix, endUnix).Scan(&rows).Error
		if err != nil {
			return fmt.Errorf("query token_transactions for %s: %w", date, err)
		}

		for _, row := range rows {
			// upsert: 按 org + user + model + date 写入
			var existing model.EnterpriseUsageStat
			result := model.DB.Where("org_id = ? AND user_id = ? AND model_name = ? AND date = ?",
				row.OrgId, row.UserId, row.ModelName, date).First(&existing)

			if result.Error != nil {
				// 不存在 → 插入
				stat := model.EnterpriseUsageStat{
					OrgId:        row.OrgId,
					DepartmentId: 0, // 通过 User.department_id 可回填
					UserId:       row.UserId,
					Date:         date,
					TokenCount:   row.TokenCount,
					RequestCount: row.RequestCount,
					Cost:         row.CostCents,
					ModelName:    row.ModelName,
				}
				if err := model.DB.Create(&stat).Error; err != nil {
					logger.SysError(fmt.Sprintf("[EnterpriseUsage] insert failed: %v", err))
				}
			} else {
				// 已存在 → 更新(幂等)
				model.DB.Model(&existing).Updates(map[string]interface{}{
					"token_count":    row.TokenCount,
					"request_count":  row.RequestCount,
					"cost":           row.CostCents,
				})
			}
		}

		// 再聚合部门级别：按 department_id 聚合
		type DeptAggRow struct {
			OrgId        int
			DepartmentId int
			ModelName    string
			TokenCount   int64
			RequestCount int64
			CostCents    int64
		}
		var deptRows []DeptAggRow
		model.DB.Raw(`
			SELECT
				u.organization_id as org_id,
				COALESCE(u.department_id, 0) as department_id,
				t.model_name,
				COALESCE(SUM(t.prompt_tokens + t.completion_tokens), 0) as token_count,
				COUNT(*) as request_count,
				COALESCE(SUM(t.total_amount * 100), 0) as cost_cents
			FROM token_transactions t
			JOIN users u ON u.id = t.user_id
			WHERE t.created_time >= ? AND t.created_time <= ?
				AND u.organization_id > 0 AND u.department_id > 0
			GROUP BY u.organization_id, u.department_id, t.model_name
		`, startUnix, endUnix).Scan(&deptRows)

		for _, row := range deptRows {
			var existing model.EnterpriseUsageStat
			result := model.DB.Where("org_id = ? AND department_id = ? AND user_id = 0 AND model_name = ? AND date = ?",
				row.OrgId, row.DepartmentId, row.ModelName, date).First(&existing)

			stat := model.EnterpriseUsageStat{
				OrgId:        row.OrgId,
				DepartmentId: row.DepartmentId,
				UserId:       0, // 部门汇总
				Date:         date,
				TokenCount:   row.TokenCount,
				RequestCount: row.RequestCount,
				Cost:         row.CostCents,
				ModelName:    row.ModelName,
			}
			if result.Error != nil {
				model.DB.Create(&stat)
			} else {
				model.DB.Model(&existing).Updates(map[string]interface{}{
					"token_count":   row.TokenCount,
					"request_count": row.RequestCount,
					"cost":          row.CostCents,
				})
			}
		}

		logger.SysLog(fmt.Sprintf("[EnterpriseUsage] aggregated %s: %d user-rows, %d dept-rows", date, len(rows), len(deptRows)))
	}

	// ── 部门预算预警 ──
	checkDepartmentBudgetAlerts()
	return nil
}

// checkDepartmentBudgetAlerts 检查部门预算使用率，超阈值时发通知
func checkDepartmentBudgetAlerts() {
	var departments []model.Department
	if err := model.DB.Where("status = 1 AND monthly_budget > 0 AND alert_threshold > 0").Find(&departments).Error; err != nil {
		return
	}
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthStartUnix := monthStart.Unix()

	for _, dept := range departments {
		// 查询本月该部门消费
		var totalCost int64
		model.DB.Raw(`
			SELECT COALESCE(SUM(t.total_amount * 100), 0)
			FROM token_transactions t
			JOIN users u ON u.id = t.user_id
			WHERE u.department_id = ? AND t.created_time >= ?
		`, dept.Id, monthStartUnix).Scan(&totalCost)

		if dept.MonthlyBudget <= 0 {
			continue
		}
		usagePct := float64(totalCost) / float64(dept.MonthlyBudget) * 100

		// 超过预警阈值 → 发通知给部门负责人+组织管理员
		if usagePct >= float64(dept.AlertThreshold) {
			// 先查是否已发过通知（避免重复）
			var existingCount int64
			model.DB.Model(&model.Notification{}).
				Where("type = ? AND data LIKE ? AND created_at > ?",
					"budget_alert", `%"department_id":`+fmt.Sprint(dept.Id)+`%`, monthStart).Count(&existingCount)
			if existingCount > 0 {
				continue // 本月已通知过
			}

			title := fmt.Sprintf("部门预算预警: %s 已使用 %.0f%%", dept.Name, usagePct)
			content := fmt.Sprintf("部门 %s 本月已使用预算 ¥%.2f，占月度预算 ¥%.2f 的 %.0f%%，超过预警阈值 %d%%。",
				dept.Name, float64(totalCost)/100, float64(dept.MonthlyBudget)/100, usagePct, dept.AlertThreshold)
			data := fmt.Sprintf(`{"department_id":%d,"usage_pct":%.0f,"cost":%d,"budget":%d}`,
				dept.Id, usagePct, totalCost, dept.MonthlyBudget)

			// 通知部门负责人
			if dept.HeadUserId > 0 {
				model.CreateNotification(dept.HeadUserId, "budget_alert", title, content, data)
			}
			// 通知组织管理员
			var admins []struct{ UserId int }
			model.DB.Raw("SELECT user_id FROM organization_members WHERE org_id = ? AND role = ?", dept.OrgId, "admin").Scan(&admins)
			for _, a := range admins {
				if a.UserId != dept.HeadUserId {
					model.CreateNotification(a.UserId, "budget_alert", title, content, data)
				}
			}
			logger.SysLog(fmt.Sprintf("[BudgetAlert] Dept #%d (%s): %.0f%% used, notified %d admins", dept.Id, dept.Name, usagePct, len(admins)))
		}
	}
}

func parseTimeToUnix(timeStr string) int64 {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	if err != nil {
		return 0
	}
	return t.Unix()
}
