import TextInput, { Props as TextInputProps } from "./TextInput";
import SelectInput, { Props as SelectInputProps } from "./SelectInput";
import { MultiSelect, Option } from "chakra-multiselect";

export type { TextInputProps, SelectInputProps, Option };

export default {
  Text: TextInput,
  Select: SelectInput,
  MultiSelect,
};
