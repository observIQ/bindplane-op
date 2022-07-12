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
	"encoding/json"
	"io"

	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/model"
)

// JSONPrinter logs json to an io.Writer
type JSONPrinter struct {
	writer io.Writer
	logger *zap.Logger
}

var _ Printer = (*JSONPrinter)(nil)

// NewJSONPrinter returns a new *JSONPrinter
func NewJSONPrinter(writer io.Writer, logger *zap.Logger) *JSONPrinter {
	return &JSONPrinter{
		writer: writer,
		logger: logger,
	}
}

// PrintResource prints a generic model that implements the printable interface
func (jp *JSONPrinter) PrintResource(item model.Printable) {
	if item == nil {
		return
	}
	jp.printIndentedJSONLine(item, item.PrintableKindSingular())
}

// PrintResources prints a generic model that implements the model.Printable interface
func (jp *JSONPrinter) PrintResources(list []model.Printable) {
	if len(list) == 0 {
		jp.printIndentedJSONLine(list, "?")
	} else {
		jp.printIndentedJSONLine(list, list[0].PrintableKindPlural())
	}
}

func (jp *JSONPrinter) printIndentedJSONLine(resource interface{}, resourceName string) {
	val, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		jp.logger.Error("could not marshal resource as json", zap.String("resource", resourceName))
		return
	}

	_, writeErr := jp.writer.Write(val)
	if writeErr != nil {
		jp.logger.Error("could not write resource as json", zap.String("resource", resourceName))
	}
}
