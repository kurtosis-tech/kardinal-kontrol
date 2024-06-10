import {
  Checkbox,
  FormControl,
  FormLabel,
  Radio,
  RadioGroup,
  Select,
  Stack,
  Text,
} from "@chakra-ui/react";

const Create = () => {
  return (
    <Stack direction="column" p="4" maxWidth={"800px"} spacing={6}>
      <Text fontSize="2xl" fontWeight="bold" mb="4">
        Create Service Deployment
      </Text>

      <Stack direction={"column"} maxWidth={"400px"} spacing={6}>
        <FormControl mb="4">
          <FormLabel>
            <b>Service</b>
          </FormLabel>
          <Select placeholder="Select service" value="AnalyticsService">
            <option value="AnalyticsService">AnalyticsService</option>
          </Select>
        </FormControl>

        <FormControl mb="4">
          <FormLabel>
            <b>Git Commit</b>
          </FormLabel>
          <Select
            placeholder="Select commit"
            value={"edf7194c5b56ab8b120fa68f1c453b46c0ada76f "}
          >
            <option value="edf7194c5b56ab8b120fa68f1c453b46c0ada76f ">
              edf7194c5b56ab8b120fa68f1c453b46c0ada76f{" "}
            </option>
          </Select>
        </FormControl>
      </Stack>

      <Stack direction={"row"}>
        <FormControl as="fieldset" mb="4">
          <FormLabel as="legend">
            <b>Traffic</b>
          </FormLabel>
          <Stack spacing="4">
            <Checkbox isChecked>Staging (Mirror)</Checkbox>
            <Checkbox isDisabled>Staging (Redirect) [Coming Soon]</Checkbox>
            <Checkbox isDisabled>Isolated Dev [Coming Soon]</Checkbox>
          </Stack>
        </FormControl>

        <FormControl as="fieldset" mb="4">
          <FormLabel as="legend">
            <b>Database</b>
          </FormLabel>
          <RadioGroup defaultValue="staging-read-only">
            <Stack spacing="4">
              <Radio value="staging-read-only">Staging Read Only</Radio>
              <Radio value="staging-read-write" isDisabled>
                Staging Read/Write [Coming Soon]
              </Radio>
              <Radio value="ephemeral" isDisabled>
                Ephemeral [Coming Soon]
              </Radio>
            </Stack>
          </RadioGroup>
        </FormControl>
      </Stack>
    </Stack>
  );
};

export default Create;
