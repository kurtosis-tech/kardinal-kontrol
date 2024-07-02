import ComingSoon from "@/components/ComingSoon";
import { FiLayout } from "react-icons/fi";

const Page = () => {
  return (
    <ComingSoon icon={FiLayout} title={"Dashboard is coming soon"}>
      We’re working on getting this functionality up and running. We’ll let you
      know when it’s ready for you!
    </ComingSoon>
  );
};

export default Page;
