import { useState } from 'react';
import IconButton from '@mui/material/IconButton';
import { FilterFilled } from '@ant-design/icons';
import Popover from '@mui/material/Popover';
import Stack from '@mui/material/Stack';
import { DatePicker } from '@mui/x-date-pickers';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import Autocomplete from '@mui/material/Autocomplete';
import TextField from '@mui/material/TextField';
import { usePaymentMethodNames } from "../../api/graph";
import { useBudget } from '../budget/BudgetContext';
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import Box from "@mui/material/Box";
import dayjs from "dayjs";

const headCells = [
  {
    id: "date",
    align: "left",
    disablePadding: false,
    label: "Date",
  },
  {
    id: "name",
    align: "left",
    disablePadding: true,
    label: "Name",
  },
  {
    id: "amount",
    align: "left",
    disablePadding: false,
    label: "Amount",
  },
  {
    id: "method",
    align: "left",
    disablePadding: false,
    label: "Method",
  },
  {
    id: "budget_category",
    align: "left",
    disablePadding: false,
    label: "Budget Category",
  },
  {
    id: "reward_category",
    align: "left",
    disablePadding: false,
    label: "Reward Category",
  },
  {
    id: "comment",
    align: "right",
    disablePadding: false,
    label: "Comment",
  },
];

export function RecentTransactionsTableHeader({ filters, onFilterChange }) {
  const [anchorEls, setAnchorEls] = useState({});
  const { data: paymentMethodsData } = usePaymentMethodNames();
  const { budget } = useBudget();
  const categories = budget?.expenses?.map(expense => expense.category);
  const since = filters.since ? dayjs(filters.since) : null;
  const until = filters.until ? dayjs(filters.until) : null;

  const handleFilterClick = (event, column) => {
    setAnchorEls(prev => ({ ...prev, [column]: event.currentTarget }));
  };

  const handleFilterClose = (column) => {
    setAnchorEls(prev => ({ ...prev, [column]: null }));
  };

  const handleDateChange = (type, date) => {
    if (type === 'start') {
        onFilterChange({
            ...filters,
            since: date.format("YYYY-MM-DD"),
        });
    } else {
        onFilterChange({
            ...filters,
            until: date.format("YYYY-MM-DD"),
        });
    }
  };

  const handleMethodChange = (_, value) => {
    onFilterChange({
      ...filters,
      paymentMethod: value?.id ?? null,
    });
  };

  const handleCategoryChange = (_, value) => {
    onFilterChange({
      ...filters,
      category: value,
    });
  };

  return (
    <TableHead>
      <TableRow>
        {headCells.map((headCell) => (
          <TableCell
            key={headCell.id}
            align={headCell.align}
            padding={headCell.disablePadding ? "none" : "normal"}
          >
            <Stack direction="row" alignItems="center" spacing={1}>
              {headCell.label}
              {(headCell.id === 'date' || headCell.id === 'method' || headCell.id === 'budget_category') && (
                <>
                  <IconButton size="small" onClick={(e) => handleFilterClick(e, headCell.id)}>
                    <FilterFilled fontSize="small" />
                  </IconButton>
                  <Popover
                    open={Boolean(anchorEls[headCell.id])}
                    anchorEl={anchorEls[headCell.id]}
                    onClose={() => handleFilterClose(headCell.id)}
                    anchorOrigin={{
                      vertical: 'bottom',
                      horizontal: 'left',
                    }}
                  >
                    {headCell.id === 'date' ? (
                      <LocalizationProvider dateAdapter={AdapterDayjs}>
                        <Stack p={2} spacing={2}>
                          <DatePicker
                            label="Start Date"
                            value={since}
                            onChange={(date) => handleDateChange('start', date)}
                            slotProps={{ textField: { size: 'small' } }}
                          />
                          <DatePicker
                            label="End Date"
                            value={until}
                            onChange={(date) => handleDateChange('end', date)}
                            slotProps={{ textField: { size: 'small' } }}
                          />
                        </Stack>
                      </LocalizationProvider>
                    ) : headCell.id === 'method' ? (
                      <Box p={2}>
                        <Autocomplete
                            size="small"
                            options={paymentMethodsData?.paymentMethods || []}
                            value={paymentMethodsData?.paymentMethods?.find(m => m.id === filters.paymentMethod) || null}
                            onChange={handleMethodChange}
                            getOptionLabel={(option) => option?.displayName || ''}
                            renderInput={(params) => <TextField {...params} label="Filter Method" />}
                            sx={{ width: 200 }}
                        />
                      </Box>
                    ) : (
                      <Box p={2}>
                        <Autocomplete
                          size="small"
                          options={categories || []}
                          value={filters.category}
                          onChange={handleCategoryChange}
                          renderInput={(params) =>
                            <TextField {...params} label="Filter Category" />
                          }
                          sx={{ width: 200 }}
                        />
                      </Box>
                    )}
                  </Popover>
                </>
              )}
            </Stack>
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}