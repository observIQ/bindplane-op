// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package printer

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"

	"github.com/observiq/bindplane-op/model"
)

// TablePrinter TODO(doc)
type TablePrinter struct {
	writer io.Writer
	table  *tablewriter.Table
}

var _ Printer = (*TablePrinter)(nil)

// NewTablePrinter takes an io.Writer and returns a new *TablePrinter.
func NewTablePrinter(writer io.Writer) *TablePrinter {
	table := tablewriter.NewWriter(writer)

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	return &TablePrinter{writer: writer, table: table}
}

// PrintResource prints a generic model that implements the printable interface
func (tp *TablePrinter) PrintResource(item model.Printable) {
	tp.PrintResources([]model.Printable{item})
}

// PrintResources prints a list of generic models that implements the printable interface
func (tp *TablePrinter) PrintResources(list []model.Printable) {
	if len(list) == 0 {
		fmt.Fprintln(tp.writer, "No matching resources found.")
		return
	}
	titles := list[0].PrintableFieldTitles()
	tp.Reset()
	tp.table.SetHeader(titles)
	for _, item := range list {
		tp.table.Append(model.PrintableFieldValuesForTitles(item, titles))
	}
	tp.table.Render()
}

// Reset TODO(docs)
func (tp *TablePrinter) Reset() {
	tp.table.ClearRows()
	tp.table.ClearFooter()
	tp.table.SetHeader([]string{})
}
