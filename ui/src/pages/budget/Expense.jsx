import { useState } from "react";
import { useBudget } from "./BudgetContext";
import TextField from "@mui/material/TextField";
import Grid2 from "@mui/material/Unstable_Grid2";
import { Fab, Switch, FormControlLabel } from "@mui/material";
import { DeleteOutlined } from "@ant-design/icons";
import {cloneDeep} from "lodash";

function Expense({ index, showAmount = true }) {
    const { budget, setBudget } = useBudget();
    const [errorMessage, setErrorMessage] = useState("");

    const expense = cloneDeep(budget.expenses[index]);

    let withExpense = () => {
        var b = cloneDeep(budget);
        b.expenses[index] = expense;
        return b
    }

    let changeCategory = (e) => {
        if (
            budget.expenses
                .map((e) => e.category)
                .filter((_, i) => i !== index)
                .includes(e.target.value)
        ) {
            setErrorMessage("Category already exists.");
        } else {
            setErrorMessage("");
            expense.category = e.target.value;
            setBudget(withExpense());
        }
    };

    let changeAmount = (e) => {
        expense.amount = parseFloat(e.target.value);
        setBudget(withExpense());
    };

    let toggleFixed = (e) => {
        expense.isFixed = e.target.checked;
        setBudget(withExpense());
    };

    return (
        <Grid2 container spacing={2} sx={{ justifyContent: "space-between" }}>
            <Grid2 xs={4}>
                <TextField
                    id={"expense-category-" + index}
                    label="Category"
                    variant="standard"
                    value={expense.category}
                    fullWidth={true}
                    onChange={changeCategory}
                    error={errorMessage !== ""}
                    helperText={errorMessage}
                />
            </Grid2>

            <Grid2 xs={5}>
                <TextField
                    id={"expense-amount-" + index}
                    label="Amount"
                    variant="standard"
                    value={expense.amount}
                    fullWidth={true}
                    onChange={changeAmount}
                />
            </Grid2>

            <Grid2 xs={2}>
                {!expense.isSlack && (
                    <FormControlLabel
                        id={"expense-fixed-" + index}
                        control={<Switch checked={expense.isFixed} />}
                        label="Fixed"
                        onChange={toggleFixed}
                    />
                )}
            </Grid2>

            <Grid2 sx={{ justifyContent: "flex-end" }} xs={1}>
                {!expense.isSlack && (
                    <Fab
                        color="secondary"
                        aria-label="remove"
                        size="small"
                        onClick={() => {
                            var b = cloneDeep(budget);
                            b.expenses.splice(index, 1);
                            setBudget(b);
                        }}
                    >
                        <DeleteOutlined style={{ fontSize: "1.3rem" }} />
                    </Fab>
                )}
            </Grid2>
        </Grid2>
    );
}

export default Expense;