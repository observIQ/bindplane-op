import { withNavBar } from "../../components/NavBar";
import { withRequireLogin } from "../../contexts/RequireLogin";

export const LiveTailPageComponent: React.FC = () => {
  return <>LIVETAIL</>;
};

export const LiveTailPage = withRequireLogin(withNavBar(LiveTailPageComponent));
