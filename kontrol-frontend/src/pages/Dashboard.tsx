import EmptyState from "@/components/EmptyState";
import { FiLayout } from "react-icons/fi";

const Page = () => {
  return (
    <EmptyState icon={FiLayout} buttonText="Go back" buttonTo="">
      We’re working on getting this functionality up and running. We’ll let you
      know when it’s ready for you!
    </EmptyState>
  );
};

export default Page;
