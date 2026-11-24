// material-ui
import Typography from "@mui/material/Typography";
import ManageBudget from "./ManageBudget";
import {BudgetProvider} from "./BudgetContext";

export default function BudgetDashboard() {
  return (
    <>
      <Typography variant="h5">Budget</Typography>
      <BudgetProvider>
        <ManageBudget />
      </BudgetProvider>
    </>
  );
}
