package metric

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

type Reporter struct {
}

func (r Reporter) Report() {
	file := xlsx.NewFile()
	defer func() {
		file.Save(Mgr.handler.GetReportFile())
	}()
	sheet, _ := file.AddSheet("响应时间(毫秒)")

	headerFilds := []string{
		"消息名",
		"请求数量",
		"响应数量",
		"响应率",
		"最小响应",
		"最大响应",
		"平均响应",
		"50%分位",
		"75%分位",
		"90%分位",
	}
	row := sheet.AddRow() // 表头
	for _, fd := range headerFilds {
		row.AddCell().Value = fd
	}

	// 数据
	for msgName, info := range Status() {
		if info.UpCount == 0 {
			continue
		}

		row := sheet.AddRow()
		row.AddCell().Value = msgName
		row.AddCell().SetValue(info.UpCount)
		row.AddCell().SetValue(info.DownCount)
		row.AddCell().SetValue(info.RspRate)
		row.AddCell().SetValue(info.Min)
		row.AddCell().SetValue(info.Max)
		row.AddCell().SetValue(info.Avg)
		row.AddCell().SetValue(info.Fifty)
		row.AddCell().SetValue(info.SeventyFive)
		row.AddCell().SetValue(info.Ninety)
	}
}

func (r Recorder) GetReportFile() string {
	return "report.xlsx"
}

func (r Recorder) TransMsgId(id interface{}) string {
	return fmt.Sprintf("%v", id)
}
