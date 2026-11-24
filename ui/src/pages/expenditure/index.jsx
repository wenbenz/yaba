import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";

import MainCard from "components/MainCard";
import RecentTransactionsTable from "./RecentTransactionsTable";
import ExportButton from "./ExportTransactions";

import {BudgetProvider} from "../budget/BudgetContext";
import Button from "@mui/material/Button";
import {PlusCircleFilled} from "@ant-design/icons";
import {useState} from "react";
import AddTransactionDialog from "./UploadTransactions";

export default function ExpenditureDashboard() {
  const [openDialog, setOpenDialog] = useState(false);

  return (
        <BudgetProvider>
          <Grid container rowSpacing={4.5} columnSpacing={2.75}>
            {/* row 1 */}
            <Grid item xs={12}>
              <Grid container alignItems="center" justifyContent="space-between">
                <Grid item>
                  <Typography variant="h5">Recent Transactions</Typography>
                </Grid>
                <Grid item>
                  <Stack direction={"row"} spacing={2}>
                    <Button
                        variant="contained"
                        color="primary"
                        onClick={() => setOpenDialog(true)}
                        mt={2}
                    >
                      <Stack direction="row" spacing={1} alignItems="center">
                        <PlusCircleFilled />
                        <Typography>
                          Add new transactions
                        </Typography>
                      </Stack>
                    </Button>
                    <AddTransactionDialog
                        open={openDialog}
                        onClose={() => setOpenDialog(false)}
                    />

                    <ExportButton />
                  </Stack>
                </Grid>
              </Grid>
              <MainCard sx={{mt: 2}} content={false}>
                <RecentTransactionsTable />
              </MainCard>
            </Grid>
          </Grid>
        </BudgetProvider>
  );
}
