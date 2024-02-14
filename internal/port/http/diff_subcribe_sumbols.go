package http

import (
	"sort"

	"example.com/greetings/pkg/enum"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func (c *diffController) SubscribeSymbols(ctx *gin.Context) {
	items := c.sv.GetCurrentSymbolItems()

	result := map[string]interface{}{}

	symbols := []string{}

	for _, v := range items {
		symbols = append(symbols, v.Symbol)
	}

	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i] < symbols[j]
	})

	result["symbols"] = symbols

	wsSymbols := []string{}
	conn := map[string]interface{}{}

	for id, v := range c.sv.GetConnections() {
		wsSymbols = append(wsSymbols, v.Symbols...)
		conn[id] = map[string]interface{}{
			"type":    enum.ExchangeTypeName[v.ExchangeType],
			"trading": enum.TradingTypeName[v.TradingType],
			"symbols": v.Symbols,
		}
	}

	result["connections"] = conn

	add, remove := lo.Difference(wsSymbols, symbols)

	result["add"] = add
	result["remove"] = remove

	wsSymbols = lo.Uniq(wsSymbols)

	sort.Slice(wsSymbols, func(i, j int) bool {
		return wsSymbols[i] < wsSymbols[j]
	})

	result["ws_symbols"] = lo.Uniq(wsSymbols)

	result["count_symbols"] = len(symbols)
	result["count_ws_symbols"] = len(wsSymbols)
	result["items"] = items

	ctx.JSON(200, result)
}
