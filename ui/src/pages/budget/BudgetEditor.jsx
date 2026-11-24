import { useBudget } from "./BudgetContext";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import { Fab } from "@mui/material";
import { PlusOutlined } from "@ant-design/icons";
import Grid2 from "@mui/material/Unstable_Grid2";
import Income from "./Income";
import Expense from "./Expense";
import {cloneDeep} from "lodash";

function SaveStatusLabel({ saveStatus }) {
    if (saveStatus === "saving") {
        return <Typography>Saving…</Typography>;
    }
    if (saveStatus === "saved") {
        return <Typography>Saved</Typography>;
    }
    if (saveStatus === "failed") {
        return <Typography color="error">Save failed</Typography>;
    }
    return null;
}

export default function BudgetEditor() {
    const { budget, setBudget, saveStatus } = useBudget();

    let totalIncome = budget.incomes.reduce((a, b) => a + b.amount, 0);
    let totalExpenses = budget.expenses.reduce((a, b) => a + b.amount, 0);

    return (
        <Stack spacing={5}>
            <TextField
                id="budgetName"
                label="Budget Name"
                variant="standard"
                fullWidth={true}
                value={budget.name}
                onChange={(e) => {
                    let b = cloneDeep(budget)
                    b.name = e.target.value;
                    setBudget(b);
                }}
            />

            <Grid2>
                <Typography variant="subtitle1">Incomes</Typography>
                <Stack spacing={2}>
                    {budget.incomes.map((_, index) => (
                        <Income key={index} index={index} />
                    ))}

                    <Grid2>
                        <Fab
                            color="primary"
                            aria-label="add"
                            size="small"
                            onClick={() => {
                                var b = cloneDeep(budget);
                                b.incomes.push({ source: "", amount: 0 });
                                setBudget(b);
                            }}
                        >
                            <PlusOutlined style={{ fontSize: "1.3rem" }} />
                        </Fab>
                    </Grid2>
                </Stack>
            </Grid2>

            <Stack spacing={4} direction="row">
                <Typography variant="subtitle2">Total Income: </Typography>
                <Typography variant="subtitle1" color="primary">
                    ${totalIncome.toFixed(2)}
                </Typography>
            </Stack>

            <Grid2>
                <Typography variant="subtitle1">Expenses</Typography>
                <Stack spacing={2}>
                    {budget.expenses
                        .map((e, index) => [
                            e.isSlack,
                            <Expense key={index} index={index} />,
                        ])
                        .filter((e) => !e[0])}

                    <Grid2>
                        <Fab
                            color="primary"
                            aria-label="add"
                            size="small"
                            onClick={() => {
                                var b = cloneDeep(budget);
                                b.expenses.push({
                                    category: "",
                                    amount: 0,
                                    isFixed: false,
                                });
                                setBudget(b);
                            }}
                        >
                            <PlusOutlined style={{ fontSize: "1.3rem" }} />
                        </Fab>
                    </Grid2>

                    <Typography variant={"subtitle2"}>All other expenses</Typography>
                    {budget.expenses
                        .map((e, index) => [
                            e.isSlack,
                            <Expense key={index} index={index} />,
                        ])
                        .filter((e) => e[0])}
                </Stack>
            </Grid2>

            <Stack spacing={4} direction="row">
                <Typography variant="subtitle2">Total Expenses: </Typography>
                <Typography variant="subtitle1" color="primary">
                    ${totalExpenses.toFixed(2)}
                </Typography>
            </Stack>

            <Stack spacing={4} direction="row">
                <Typography variant="subtitle2">Extra: </Typography>
                <Typography
                    variant="subtitle1"
                    color={totalIncome >= totalExpenses ? "primary" : "error"}
                >
                    ${(totalIncome - totalExpenses).toFixed(2)}
                </Typography>
            </Stack>

            <SaveStatusLabel saveStatus={saveStatus} />
        </Stack>
    );
}