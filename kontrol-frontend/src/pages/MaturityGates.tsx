import EmptyState from "@/components/EmptyState";
import { BiChevronLeft } from "react-icons/bi";
import { FiShield } from "react-icons/fi";

const Page = () => {
  return (
    <EmptyState
      icon={FiShield}
      buttonText="Go back"
      buttonIcon={BiChevronLeft}
      buttonTo="../traffic-configuration"
    >
      We’re working on getting this functionality up and running. We’ll let you
      know when it’s ready for you!
    </EmptyState>
  );
};

export default Page;
