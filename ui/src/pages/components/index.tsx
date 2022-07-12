import { CardContainer } from "../../components/CardContainer";
import { withNavBar } from "../../components/NavBar";
import { ComponentsTable } from "../../components/Tables/ComponentsTable";
import { withRequireLogin } from "../../contexts/RequireLogin";

const ComponentsPageContent: React.FC = () => {
  return (
    <>
      <CardContainer>
        <ComponentsTable />
      </CardContainer>
    </>
  );
};

export const ComponentsPage = withRequireLogin(
  withNavBar(ComponentsPageContent)
);
