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

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/observiq/bindplane-op/model"
)

// YamlPrinter TODO(doc)
type YamlPrinter struct {
	writer io.Writer
	logger *zap.Logger
}

var _ Printer = (*YamlPrinter)(nil)

// NewYamlPrinter TODO(doc)
func NewYamlPrinter(writer io.Writer, logger *zap.Logger) *YamlPrinter {
	return &YamlPrinter{
		writer: writer,
		logger: logger,
	}
}

// PrintResource prints a generic model that implements the printable interface
func (yp *YamlPrinter) PrintResource(item model.Printable) {
	if item == nil {
		return
	}
	yp.printYamlLine(item, item.PrintableKindSingular())
}

// PrintResources prints a generic model that implements the model.Printable interface
func (yp *YamlPrinter) PrintResources(list []model.Printable) {
	if len(list) == 0 {
		fmt.Fprintln(yp.writer, "[]")
		return
	}
	for _, item := range list {
		fmt.Fprintln(yp.writer, "---")
		yp.printYamlLine(item, list[0].PrintableKindPlural())
	}
}

func (yp *YamlPrinter) printYamlLine(resource interface{}, resourceName string) {
	val, err := yaml.Marshal(resource)
	if err != nil {
		yp.logger.Error("could marshal resource as yaml", zap.String("resource", resourceName))
		return
	}

	_, writeErr := yp.writer.Write(val)
	if writeErr != nil {
		yp.logger.Error("could not write resource as yaml", zap.String("resource", resourceName))
	}
}
