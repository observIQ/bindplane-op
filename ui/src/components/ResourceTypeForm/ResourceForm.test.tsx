import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { ResourceConfigForm } from ".";
import {
  ParameterDefinition,
  ParameterType,
  RelevantIfOperatorType,
} from "../../graphql/generated";
import {
  ParameterInput,
  Tuple,
  tupleArrayToMap,
  valueToTupleArray,
} from "./ParameterInput";
import { satisfiesRelevantIf } from "./satisfiesRelevantIf";
import { ResourceType1, ResourceType2 } from "./__test__/dummyResources";
import renderer from "react-test-renderer";
import { type } from "os";

describe("satisfiesRelevantIf", () => {
  const formValues: { [key: string]: any } = {
    one: true,
    two: "bar",
    three: 25,
  };

  const param1: ParameterDefinition = {
    name: "string_name",
    label: "String Input",
    description: "Here is the description.",
    required: false,

    type: ParameterType.String,

    relevantIf: [
      {
        name: "one",
        operator: RelevantIfOperatorType.Equals,
        value: true,
      },
    ],
  };

  const param2: ParameterDefinition = {
    name: "string_name",
    label: "String Input",
    description: "Here is the description.",
    required: false,

    type: ParameterType.String,
    relevantIf: [
      {
        name: "one",
        operator: RelevantIfOperatorType.Equals,
        value: false,
      },
    ],

    default: "default-value",
  };

  it("param1 matches", () => {
    const got = satisfiesRelevantIf(formValues, param1);
    expect(got).toEqual(true);
  });
  it("param2 does not match", () => {
    const got = satisfiesRelevantIf(formValues, param2);
    expect(got).toEqual(false);
  });
});

describe("ResourceForm component", () => {
  it("does not display field if relevantIf isn't satisfied", () => {
    render(
      <ResourceConfigForm
        kind="destination"
        title={ResourceType2.metadata.displayName!}
        description={ResourceType2.metadata.description!}
        parameterDefinitions={ResourceType2.spec.parameters}
      />
    );
    const stringInput = screen.queryByText("String Input");
    expect(stringInput).toBeNull();
  });

  it("will render input when relevantIf is satisfied", () => {
    render(
      <ResourceConfigForm
        kind="destination"
        title={ResourceType2.metadata.displayName!}
        description={ResourceType2.metadata.description!}
        parameterDefinitions={ResourceType2.spec.parameters}
      />
    );
    let stringInput = screen.queryByLabelText("String Input");
    expect(stringInput).toBeNull();

    screen.getByRole("checkbox").click();
    stringInput = screen.getByLabelText("String Input");
    expect(stringInput).toBeInTheDocument();
  });

  it("maintains stateful formValues as correctType", async () => {
    const expectedValues = {
      name: "",
      string_name: "default-value",
      string_required_name: "default-required-value",
      enum_name: "option1",
      strings_name: ["option1", "option2"],
      int_name: 25,
      bool_name: true,
    };

    let saveDone = false;

    let values: { [key: string]: any } = {};
    function onSave(formValues: { [key: string]: any }) {
      values = Object.assign({}, formValues);
      saveDone = true;
    }
    render(
      <ResourceConfigForm
        onSave={onSave}
        kind="source"
        title={ResourceType1.metadata.displayName!}
        description={ResourceType1.metadata.description!}
        parameterDefinitions={ResourceType1.spec.parameters}
        includeNameField
      />
    );

    screen.getByText("Save").click();

    await waitFor(() => saveDone === true);
    expect(values).toEqual(expectedValues);
  });

  it("maintains stateful formValues as correctType after change", async () => {
    const expectedValues = {
      name: "",
      string_name: "default-value",
      string_required_name: "default-required-value",
      enum_name: "option1",
      strings_name: ["option1", "option2"],
      int_name: 50,
      bool_name: true,
    };

    let saveDone = false;

    let values: { [key: string]: any } = {};
    function onSave(formValues: { [key: string]: any }) {
      values = Object.assign({}, formValues);
      saveDone = true;
    }
    render(
      <ResourceConfigForm
        onSave={onSave}
        kind="source"
        title={ResourceType1.metadata.displayName!}
        description={ResourceType1.metadata.description!}
        parameterDefinitions={ResourceType1.spec.parameters}
        includeNameField
      />
    );

    fireEvent.change(screen.getByLabelText("Int Input"), {
      target: { value: 50 },
    });

    screen.getByText("Save").click();

    await waitFor(() => saveDone === true);
    expect(values).toEqual(expectedValues);
  });

  it("disables save button when name field has an error", async () => {
    render(
      <ResourceConfigForm
        onSave={() => {}}
        kind="destination"
        title={ResourceType1.metadata.displayName!}
        description={ResourceType1.metadata.description!}
        parameterDefinitions={ResourceType1.spec.parameters}
        includeNameField
      />
    );

    const nameField = screen.getByTestId("name-field");
    // this is an invalid name and should make the save button disabled
    fireEvent.change(nameField, { target: { value: "dest-" } });

    expect(screen.getByTestId("resource-form-save")).toBeDisabled();
  });

  it("strings type parameter validation", async () => {
    const label = "p1 label";
    const definitions: ParameterDefinition[] = [
      {
        name: "p1",
        label: label,
        type: ParameterType.Strings,
        required: true,
        description: "",
      },
    ];

    render(
      <ResourceConfigForm
        onSave={() => {}}
        kind="destination"
        title={"Title"}
        description={ResourceType1.metadata.description!}
        parameterDefinitions={definitions}
      />
    );

    expect(screen.getByTestId("resource-form-save")).toBeDisabled();

    const input = screen.getByLabelText(`${label} *`);
    fireEvent.change(input, { target: { value: "/tmp.file.log" } });
    fireEvent.keyDown(input, { key: "Enter", code: "Enter", charCode: 13 });

    expect(screen.getByTestId("resource-form-save")).not.toBeDisabled();
  });
});

describe("MapParamInput", () => {
  it("valueToTupleArray", () => {
    const tests = [
      {
        value: {
          foo: "bar",
          blah: "baz",
        },
        expect: [
          ["foo", "bar"],
          ["blah", "baz"],
          ["", ""],
        ],
      },
      {
        value: null,
        expect: [["", ""]],
      },
    ];

    for (const test of tests) {
      const got = valueToTupleArray(test.value);
      expect(got).toEqual(test.expect);
    }
  });

  it("tupleArrayToMap", () => {
    const tests: { tuples: Tuple[]; expect: any }[] = [
      {
        tuples: [
          ["one", "two"],
          ["three", "four"],
        ],
        expect: {
          one: "two",
          three: "four",
        },
      },
      {
        tuples: [
          ["", "blah"],
          ["three", "four"],
          ["some", "thing"],
          ["", ""],
        ],
        expect: {
          three: "four",
          some: "thing",
        },
      },
      {
        tuples: [["", ""]],
        expect: {},
      },
    ];

    for (const test of tests) {
      const got = tupleArrayToMap(test.tuples);
      expect(got).toEqual(test.expect);
    }
  });

  it("renders correctly", () => {
    const mapParameter: ParameterDefinition = {
      required: true,
      label: "Label",
      description: "description",
      type: ParameterType.Map,
      default: {},
      name: "map_type_param",
    };
    const tree = renderer.create(<ParameterInput definition={mapParameter} />);
    expect(tree).toMatchSnapshot();
  });
});

describe("EnumsParameter", () => {
  it("renders correctly", () => {
    const enumsParameter: ParameterDefinition = {
      required: true,
      label: "Label",
      description: "description",
      type: ParameterType.Enums,
      default: {},
      validValues: ["one", "two", "three", "four"],
      name: "enums_type_param",
    };

    const tree = renderer.create(
      <ParameterInput definition={enumsParameter} />
    );
    expect(tree).toMatchSnapshot();
  });
});

describe("YamlParameter", () => {
  it("renders correctly", () => {
    const yamlParameter: ParameterDefinition = {
      required: true,
      label: "Label",
      description: "description",
      type: ParameterType.Yaml,
      default: "",
      name: "yaml_type_param",
    };

    const tree = renderer.create(<ParameterInput definition={yamlParameter} />);
    expect(tree).toMatchSnapshot();
  });
});
