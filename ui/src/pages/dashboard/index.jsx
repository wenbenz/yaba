import Grid from "@mui/material/Grid";
import Typography from "@mui/material/Typography";
import { useState } from "react";
import {BudgetProvider} from "../budget/BudgetContext";
import {DatePicker, LocalizationProvider} from "@mui/x-date-pickers";
import {AdapterDayjs} from "@mui/x-date-pickers/AdapterDayjs";
import Stack from "@mui/material/Stack";
import dayjs from "dayjs";
import {DateRangeProvider, useDateRange} from "../../components/DateRangeProvider";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import MainCard from "../../components/MainCard";
import Box from "@mui/material/Box";
import SpendingChart from "./SpendingChart";
import CategoryChart from "./CategoryChart";

const DateRangeSelector = () => {
    const { startDate, setStartDate, endDate, setEndDate } = useDateRange();
    const [selectedRange, setSelectedRange] = useState('ytd');
    const [customEnabled, setCustomEnabled] = useState(false);

    const handleRangeChange = (event) => {
        const value = event.target.value;
        setSelectedRange(value);

        const now = new Date();
        let start = now;
        let end = now;

        switch (value) {
            case 'thisMonth':
                start = dayjs().startOf('month').toDate();
                end = dayjs().endOf('month').toDate();
                setCustomEnabled(false);
                break;
            case 'lastMonth':
                start = dayjs().subtract(1, 'month').startOf('month').toDate();
                end = dayjs().subtract(1, 'month').endOf('month').toDate();
                setCustomEnabled(false);
                break;
            case 'last3':
                start = dayjs().subtract(3, 'month').startOf('month').toDate();
                end = dayjs().endOf('month').toDate();
                setCustomEnabled(false);
                break;
            case 'ytd':
                start = dayjs().startOf('year').toDate();
                end = now;
                setCustomEnabled(false);
                break;
            case 'past12':
                start = dayjs().subtract(11, 'month').startOf('month').toDate();
                end = dayjs().endOf('month').toDate();
                setCustomEnabled(false);
                break;
            default:
                setCustomEnabled(true);
                return;
        }

        setStartDate(start);
        setEndDate(end);
    };

    return (
            <LocalizationProvider dateAdapter={AdapterDayjs}>
                <Stack direction="row" spacing={2} alignItems="center">
                    <Select
                        size="small"
                        value={selectedRange}
                        onChange={handleRangeChange}
                        sx={{ minWidth: 150 }}
                    >
                        <MenuItem value="thisMonth">This Month</MenuItem>
                        <MenuItem value="lastMonth">Last Month</MenuItem>
                        <MenuItem value="last3">Last 3 Months</MenuItem>
                        <MenuItem value="ytd">Year to Date</MenuItem>
                        <MenuItem value="past12">Past 12 Months</MenuItem>
                        <MenuItem value="custom">Custom Range</MenuItem>
                    </Select>
                    <DatePicker
                        label="Since"
                        value={dayjs(startDate)}
                        onChange={date => setStartDate(date.toDate())}
                        disabled={!customEnabled}
                        slotProps={{ textField: { size: 'small' } }}
                    />
                    <DatePicker
                        label="Until"
                        value={dayjs(endDate)}
                        onChange={date => setEndDate(date.toDate())}
                        disabled={!customEnabled}
                        slotProps={{ textField: { size: 'small' } }}
                    />
                </Stack>
            </LocalizationProvider>
    );
};

// DashboardDefault.jsx
export default function DashboardDefault() {
  return (
    <DateRangeProvider>
      <BudgetProvider>
        <Grid container rowSpacing={4.5} columnSpacing={2.75}>
          <Grid container item xs={12} sx={{ mb: -2.25 }}>
            <Grid item xs={6}>
              <Typography variant="h5">Overview</Typography>
            </Grid>
            <Grid item xs={6} sx={{ display: 'flex', justifyContent: 'flex-end' }}>
              <DateRangeSelector />
            </Grid>
          </Grid>

          <Grid item xs={12} md={6}>
            <Typography variant="h5">Budget vs Spending</Typography>
            <MainCard content={false} sx={{ mt: 1.5, height: 450 }}>
              <Box sx={{ p: 3 }}>
                <SpendingChart />
              </Box>
            </MainCard>
          </Grid>

          <Grid item xs={12} md={6}>
            <Typography variant="h5">Spending by Category</Typography>
            <MainCard content={false} sx={{ mt: 1.5, height: 450 }}>
              <Box sx={{ p: 3 }}>
                <CategoryChart />
              </Box>
            </MainCard>
          </Grid>
        </Grid>
      </BudgetProvider>
    </DateRangeProvider>
  );
}