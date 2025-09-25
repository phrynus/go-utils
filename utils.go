// Package utils 导出全部方法
package utils

import (
	"github.com/phrynus/go-utils/dingtalk"
	"github.com/phrynus/go-utils/logger"
	"github.com/phrynus/go-utils/ta"
)

// 导出logger和ta包的所有功能

// ========== Logger 包导出 ==========
const (
	INFO  = logger.INFO  // 信息级别
	DEBUG = logger.DEBUG // 调试级别
	WARN  = logger.WARN  // 警告级别
	ERROR = logger.ERROR // 错误级别
)

type LogConfig = logger.LogConfig // 日志配置
type Logger = logger.Logger       // 日志记录器

var NewLogger = logger.NewLogger // 创建日志记录器

// ========== TA 包导出 ==========
type KlineData = ta.KlineData                       // K线数据
type KlineDatas = ta.KlineDatas                     // K线数据集
type TaADX = ta.TaADX                               // 平均趋向指标
type TaATR = ta.TaATR                               // 平均真实范围指标
type TaBoll = ta.TaBoll                             // 布林带指标
type TaCCI = ta.TaCCI                               // 顺势指标
type TaCMF = ta.TaCMF                               // 能量潮指标
type TaDpo = ta.TaDpo                               // 派克指标
type TaEMA = ta.TaEMA                               // 指数移动平均线指标
type TaJingZheMA = ta.TaJingZheMA                   // 精折线指标
type TaKDJ = ta.TaKDJ                               // 随机指标
type TaMacd = ta.TaMacd                             // 移动平均线收敛/发散指标
type TaOBV = ta.TaOBV                               // 能量潮指标
type TaRMA = ta.TaRMA                               // 相对移动平均线指标
type TaRSI = ta.TaRSI                               // 相对强弱指数指标
type TaSMA = ta.TaSMA                               // 简单移动平均线指标
type TaStochRSI = ta.TaStochRSI                     // 随机相对强弱指数指标
type TaSuperTrend = ta.TaSuperTrend                 // 超级趋势指标
type TaSuperTrendPivot = ta.TaSuperTrendPivot       // 超级趋势指标
type TaSuperTrendPivotHl2 = ta.TaSuperTrendPivotHl2 // 超级趋势指标
type TaT3 = ta.TaT3                                 // 三重指数移动平均线指标
type TaVolatilityRatio = ta.TaVolatilityRatio       // 波动率比率指标
type TaWilliamsR = ta.TaWilliamsR                   // 威廉指标

var NewKlineDatas = ta.NewKlineDatas // 创建K线数据集

// ========== Dingtalk 包导出 ==========
const (
	MsgTypeText       = dingtalk.MsgTypeText       // 文本消息类型
	MsgTypeLink       = dingtalk.MsgTypeLink       // 链接消息类型
	MsgTypeMarkdown   = dingtalk.MsgTypeMarkdown   // markdown消息类型
	MsgTypeFeedCard   = dingtalk.MsgTypeFeedCard   // FeedCard消息类型
	MsgTypeActionCard = dingtalk.MsgTypeActionCard // ActionCard消息类型
)

type ResponseMeta = dingtalk.ResponseMeta                 // 响应操作信息
type DingTalk = dingtalk.DingTalk                         // 钉钉客户端
type Message = dingtalk.Message                           // 钉钉自定义机器人消息
type TextMeta = dingtalk.TextMeta                         // 文本消息
type AtMeta = dingtalk.AtMeta                             // @用户
type LinkMeta = dingtalk.LinkMeta                         // 链接消息
type MarkdownMeta = dingtalk.MarkdownMeta                 // markdown消息
type SingleActionCardMeta = dingtalk.SingleActionCardMeta // 整体跳转ActionCard
type ActionCardMeta = dingtalk.ActionCardMeta             // 独立跳转ActionCard
type ActionCardBtnMeta = dingtalk.ActionCardBtnMeta       // 按钮
type FeedCardMeta = dingtalk.FeedCardMeta                 // FeedCard消息
type FeedCardLinkMeta = dingtalk.FeedCardLinkMeta         // 链接信息

var NewDingtalk = dingtalk.NewDingtalk // 创建新的钉钉客户端
