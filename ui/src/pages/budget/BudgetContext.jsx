import {createContext, useContext, useEffect, useRef, useState} from "react";
import {useBudgets, useUpdateBudget} from "../../api/graph";
import Typography from "@mui/material/Typography";
import Loader from "../../components/Loader";
import CreateBudget from "./CreateBudget";
import debounce from "lodash.debounce";

const BudgetContext = createContext();

const templateBudget = {
    name: "My Budget",
    incomes: [
        {
            source: "Work",
            amount: 2400.0,
        },
    ],
    expenses: [
        {
            category: "Rent",
            amount: 1000.0,
            isFixed: true,
            isSlack: false,
        },
        {
            category: "Groceries",
            amount: 600.0,
            isFixed: false,
            isSlack: false,
        },
        {
            category: "Miscellaneous",
            amount: 800.0,
            isFixed: false,
            isSlack: true,
        },
    ],
};

export function BudgetProvider({ children }) {
    const { loading, data, error } = useBudgets(1);
    const [budget, setBudget] = useState(templateBudget);
    const [saveBudget] = useUpdateBudget(budget);
    const [saveStatus, setSaveStatus] = useState("idle");
    const skipNextSaveRef = useRef(true);
    const savedClearTimerRef = useRef(null);
    const hasBudget = data && data.budgets && data.budgets.length > 0;

    useEffect(() => {
        if (hasBudget) {
            setBudget(data.budgets[0]);
        }
    }, [data, hasBudget]);

    useEffect(() => {
        if (!budget.id) {
            return;
        }

        if (skipNextSaveRef.current) {
            skipNextSaveRef.current = false;
            return;
        }

        const debouncedSave = debounce(async () => {
            clearTimeout(savedClearTimerRef.current);
            setSaveStatus("saving");
            try {
                await saveBudget();
                setSaveStatus("saved");
                savedClearTimerRef.current = setTimeout(() => setSaveStatus("idle"), 2000);
            } catch {
                setSaveStatus("failed");
            }
        }, 1000);

        debouncedSave();

        return () => {
            debouncedSave.cancel();
            clearTimeout(savedClearTimerRef.current);
        };
    }, [budget]);

    if (loading) {
        return <Loader />;
    }

    if (error) {
        return (
            <Typography variant="h5">An unexpected error has occurred.</Typography>
        );
    }

    if (!hasBudget){
        return <CreateBudget budget={templateBudget} />
    }

    return (
        <BudgetContext.Provider value={{ budget, setBudget, saveStatus }}>
            {children}
        </BudgetContext.Provider>
    );
}

export function useBudget() {
    return useContext(BudgetContext);
}
