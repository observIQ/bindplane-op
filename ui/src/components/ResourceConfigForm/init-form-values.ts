import { FormValues } from ".";
import {
  ParameterDefinition,
  Parameter,
  ResourceConfiguration,
} from "../../graphql/generated";

export function initFormValues(
  definitions: ParameterDefinition[],
  parameters?: Parameter[] | null,
  processors?: ResourceConfiguration[] | null,
  includeNameField?: boolean
): FormValues {
  // Assign defaults
  let defaults: FormValues = {};
  if (includeNameField) {
    defaults.name = "";
  }

  for (const definition of definitions) {
    defaults[definition.name] = definition.default;
  }

  // Override with existing values if present
  if (parameters != null) {
    for (const parameter of parameters) {
      defaults[parameter.name] = parameter.value;
    }
  }

  // Set the processors value if present
  if (processors != null) {
    defaults.processors = processors;
  }
  return defaults;
}
