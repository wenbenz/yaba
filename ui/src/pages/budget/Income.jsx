import { useState } from "react";
import { useBudget } from "./BudgetContext";
import TextField from "@mui/material/TextField";
import Grid2 from "@mui/material/Unstable_Grid2";
import { Fab } from "@mui/material";
import { DeleteOutlined } from "@ant-design/icons";
import {cloneDeep} from "lodash";

function Income({ index }) {
    const { budget, setBudget } = useBudget();
    const [errorMessage, setErrorMessage] = useState("");

    const income = cloneDeep(budget.incomes[index]);

    let withIncome = () => {
        var b = cloneDeep(budget);
        b.incomes[index] = income;
        return b
    }

    let changeSource = (e) => {
        if (
            budget.incomes
                .map((i) => i.source)
                .filter((_, i) => i !== index)
                .includes(e.target.value)
        ) {
            setErrorMessage("Source already exists.");
        } else {
            income.source = e.target.value;
            setErrorMessage("");
            setBudget(withIncome());
        }
    };

    let changeAmount = (e) => {
        income.amount = parseFloat(e.target.value);
        setBudget(withIncome());
    };

    return (
        <Grid2 container spacing={2} sx={{ justifyContent: "space-between" }}>
            <Grid2 xs={5}>
                <TextField
                    id={"income-source-" + index}
                    label="Source"
                    variant="standard"
                    value={income.source}
                    fullWidth={true}
                    onChange={changeSource}
                    error={errorMessage !== ""}
                    helperText={errorMessage}
                />
            </Grid2>

            <Grid2 xs={6}>
                <TextField
                    id={"income-amount-" + index}
                    label="Amount"
                    variant="standard"
                    value={income.amount}
                    fullWidth={true}
                    onChange={changeAmount}
                />
            </Grid2>

            <Grid2 sx={{ justifyContent: "flex-end" }} xs={1}>
                <Fab
                    color="secondary"
                    aria-label="remove"
                    size="small"
                    onClick={() => {
                        var b = cloneDeep(budget);
                        b.incomes.splice(index, 1);
                        setBudget(b);
                    }}
                >
                    <DeleteOutlined style={{ fontSize: "1.3rem" }} />
                </Fab>
            </Grid2>
        </Grid2>
    );
}

export default Income;