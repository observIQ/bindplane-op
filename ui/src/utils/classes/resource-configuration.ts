import {
  Maybe,
  Parameter,
  ResourceConfiguration,
} from "../../graphql/generated";

export class BPResourceConfiguration implements ResourceConfiguration {
  name?: Maybe<string> | undefined;
  type?: Maybe<string> | undefined;
  parameters?: Maybe<Parameter[]> | undefined;
  processors?: Maybe<ResourceConfiguration[]> | undefined;
  constructor(rc?: ResourceConfiguration) {
    this.name = rc?.name;
    this.type = rc?.type;
    this.parameters = rc?.parameters;
    this.processors = rc?.processors;
  }

  isInline(): boolean {
    return this.name == null;
  }

  hasConfigurationParameters(): boolean {
    return this.parameters != null && this.parameters.length > 0;
  }

  // setParamsFromMap will set the parameters from Record<string, any>.
  // If the "name" key is specified it will set the name field of the ResourceConfiguration.
  // If the "processors" key is specified it will set the processors value.
  setParamsFromMap(map: Record<string, any>) {
    if (map.name != null && map.name !== "") {
      this.name = map.name;
      delete map.name;
    }

    if (map.processors != null) {
      this.processors = map.processors;
      delete map.processors;
    }

    this.parameters = Object.entries(map).map(([k, v]) => ({
      name: k,
      value: v,
    }));
  }
}
