package model

import (
	"context"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/utils"
)

type Ability struct {
	Group     string `json:"group" gorm:"type:varchar(32);primaryKey;autoIncrement:false"`
	Model     string `json:"model" gorm:"primaryKey;autoIncrement:false"`
	ChannelId int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	UserId    int    `json:"user_id" gorm:"default:0;index"`
	Enabled   bool   `json:"enabled"`
	Priority  *int64 `json:"priority" gorm:"bigint;default:0;index"`
}

func GetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*Channel, error) {
	ability := Ability{}
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	var err error = nil
	var channelQuery *gorm.DB
	if ignoreFirstPriority {
		channelQuery = DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
	} else {
		maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
		channelQuery = DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and priority = (?)", group, model, maxPrioritySubQuery)
	}
	if common.UsingSQLite || common.UsingPostgreSQL {
		err = channelQuery.Order("RANDOM()").First(&ability).Error
	} else {
		err = channelQuery.Order("RAND()").First(&ability).Error
	}
	if err != nil {
		return nil, err
	}
	channel := Channel{}
	channel.Id = ability.ChannelId
	err = DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	models_ = utils.DeDuplication(models_)
	groups_ := strings.Split(channel.Group, ",")
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				UserId:    channel.UserId,
				Enabled:   channel.Status == ChannelStatusEnabled,
				Priority:  channel.Priority,
			}
			abilities = append(abilities, ability)
		}
	}
	return DB.Create(&abilities).Error
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	// A quick and dirty way to update abilities
	// First delete all abilities of this channel
	err := channel.DeleteAbilities()
	if err != nil {
		return err
	}
	// Then add new abilities
	err = channel.AddAbilities()
	if err != nil {
		return err
	}
	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func GetGroupModels(ctx context.Context, group string) ([]string, error) {
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	var models []string
	err := DB.Model(&Ability{}).Distinct("model").Where(groupCol+" = ? and enabled = "+trueVal, group).Pluck("model", &models).Error
	if err != nil {
		return nil, err
	}
	sort.Strings(models)
	return models, err
}

// GetRandomSatisfiedChannelByOwner 按 group + model + ownerId 查 channel
// ownerId=0 → 平台渠道, ownerId>0 → 指定渠道商
func GetRandomSatisfiedChannelByOwner(group string, model string, ownerId int) (*Channel, error) {
	ability := Ability{}
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	// 取最高 priority 的 channel（限 owner）
	maxPrioritySubQuery := DB.Model(&Ability{}).
		Select("MAX(priority)").
		Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and user_id = ?", group, model, ownerId)
	channelQuery := DB.
		Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and user_id = ? and priority = (?)",
			group, model, ownerId, maxPrioritySubQuery)

	if common.UsingSQLite || common.UsingPostgreSQL {
		channelQuery.Order("RANDOM()").First(&ability)
	} else {
		channelQuery.Order("RAND()").First(&ability)
	}
	if ability.ChannelId == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	channel := Channel{}
	channel.Id = ability.ChannelId
	err := DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}

// GetRandomSatisfiedChannelAnyOwner 全资源池兜底——不限 UserId
func GetRandomSatisfiedChannelAnyOwner(group string, model string, excludeOwnerId int) (*Channel, error) {
	ability := Ability{}
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	query := DB.
		Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)

	// 排除已经试过的 owner（避免死循环）
	if excludeOwnerId >= 0 {
		query = query.Where("user_id != ?", excludeOwnerId)
	}

	// 取最高 priority
	maxPrioritySubQuery := DB.Model(&Ability{}).
		Select("MAX(priority)").
		Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
	query = query.Where("priority = (?)", maxPrioritySubQuery)

	if common.UsingSQLite || common.UsingPostgreSQL {
		query.Order("RANDOM()").First(&ability)
	} else {
		query.Order("RAND()").First(&ability)
	}
	if ability.ChannelId == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	channel := Channel{}
	channel.Id = ability.ChannelId
	err := DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}
