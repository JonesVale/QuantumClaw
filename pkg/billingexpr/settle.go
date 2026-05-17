package billingexpr

import (
	"github.com/expr-lang/expr/vm"
)

type SettleResult struct {
	Quota       int
	RefundQuota int
	Message     string
}

func CalculateQuota(exprStr string, ctx *BillingContext) (int, error) {
	if exprStr == "" {
		return 0, nil
	}

	prog, err := CompileFromCache(exprStr)
	if err != nil {
		return 0, err
	}

	vars := RuntimeVars{
		P:     ctx.PromptTokens,
		C:     ctx.CompletionTokens,
		Len:   ctx.TotalTokens,
		Cr:    ctx.ChannelRatio,
		Cc:    ctx.CustomRatio,
		Cc1h:  ctx.CustomRatio,
		Img:   ctx.ImageCount,
		ImgO:  ctx.ImageCount,
		Ai:    ctx.AudioSeconds,
		Ao:    ctx.AudioSeconds,
	}

	result, err := Run(prog, vars)
	if err != nil {
		return 0, err
	}

	return int(result), nil
}

func CompileAndExecute(exprStr string, ctx *BillingContext) (int, error) {
	prog, err := CompileFromCache(exprStr)
	if err != nil {
		return 0, err
	}

	return ExecuteCompiled(prog, ctx)
}

func ExecuteCompiled(prog *vm.Program, ctx *BillingContext) (int, error) {
	vars := RuntimeVars{
		P:     ctx.PromptTokens,
		C:     ctx.CompletionTokens,
		Len:   ctx.TotalTokens,
		Cr:    ctx.ChannelRatio,
		Cc:    ctx.CustomRatio,
		Cc1h:  ctx.CustomRatio,
		Img:   ctx.ImageCount,
		ImgO:  ctx.ImageCount,
		Ai:    ctx.AudioSeconds,
		Ao:    ctx.AudioSeconds,
	}

	result, err := Run(prog, vars)
	if err != nil {
		return 0, err
	}

	return int(result), nil
}

func SettleTask(prog *vm.Program, ctx *BillingContext, preConsumed int) SettleResult {
	result := SettleResult{
		Quota:       preConsumed,
		RefundQuota: 0,
	}

	if prog == nil {
		return result
	}

	actual, err := ExecuteCompiled(prog, ctx)
	if err != nil {
		result.Message = err.Error()
		return result
	}

	if actual > preConsumed {
		result.Quota = actual
		result.Message = "补扣"
	} else if actual < preConsumed {
		result.RefundQuota = preConsumed - actual
		result.Message = "退还"
	}

	return result
}