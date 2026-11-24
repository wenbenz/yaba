import { createContext, useContext, useState } from "react";
import { startOfYear } from "../utils/dates";

const DateRangeContext = createContext();

export const useDateRange = () => useContext(DateRangeContext);

export const DateRangeProvider = ({ children }) => {
    const [startDate, setStartDate] = useState(startOfYear());
    const [endDate, setEndDate] = useState(new Date());

    return (
        <DateRangeContext.Provider value={{ startDate, setStartDate, endDate, setEndDate }}>
            {children}
        </DateRangeContext.Provider>
    );
};
