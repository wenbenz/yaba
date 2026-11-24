import {useState, useRef} from 'react';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import IconButton from '@mui/material/IconButton';
import {DeleteOutlined, PlusOutlined} from '@ant-design/icons';
import {useBudget} from '../budget/BudgetContext';
import {UploadOutlined} from '@ant-design/icons';
import {useCreateExpenditures} from "../../api/graph";
import dayjs from "dayjs";
import {Autocomplete} from "@mui/material";
import {COMMON_REWARDS_CATEGORIES} from "../../utils/constants";
import {usePaymentMethodNames} from "../../api/graph";
import {PaymentMethodDialog} from "./PaymentMethodDialogue";

export default function AddTransactionDialog({open, onClose}) {
    const {budget} = useBudget();
    const {data: paymentMethodsData} = usePaymentMethodNames();
    const [transactions, setTransactions] = useState([{
        amount: '',
        date: new Date().toISOString().split('T')[0],
        name: '',
        method: '',
        budget_category: '',
        reward_category: '',
        comment: '',
        source: 'user'
    }]);
    const fileInputRef = useRef(null);
    const [createExpenditures] = useCreateExpenditures(transactions);
    const [isMethodDialogOpen, setIsMethodDialogOpen] = useState(false);
    const [selectedMethodId, setSelectedMethodId] = useState(null);


    const handleMethodSelect = (methodId) => {
        setIsMethodDialogOpen(false);
        setSelectedMethodId(methodId);
        fileInputRef.current.click();
    };


    const handleFileUpload = (event) => {
        const file = event.target.files[0];
        if (!file || !selectedMethodId) return;

        const reader = new FileReader();
        reader.onload = (e) => {
            const text = e.target.result;
            const lines = text.split('\n');
            const headers = lines[0].toLowerCase().split(',');

            const newTransactions = lines.slice(1)
                .filter(line => line.trim())
                .map(line => {
                    const values = line.split(',');
                    const transaction = {
                        amount: '',
                        date: new Date().toISOString().split('T')[0],
                        name: '',
                        method: selectedMethodId,
                        budget_category: '',
                        reward_category: '',
                        comment: '',
                        source: file.name,
                    };

                    headers.forEach((header, index) => {
                        const value = values[index]?.trim();
                        if (value) {
                            switch (header.trim()) {
                                case 'amount':
                                    transaction.amount = parseFloat(value) || '';
                                    break;
                                case 'date':
                                    transaction.date = dayjs(value).toISOString().split('T')[0];
                                    break;
                                case 'name':
                                case 'description':
                                    transaction.name = value;
                                    break;
                                case 'budget_category':
                                    transaction.budget_category = value;
                                    break;
                                case 'reward_category':
                                    transaction.reward_category = value.toUpperCase();
                                    break;
                                case 'comment':
                                    transaction.comment = value;
                                    break;
                                default:
                                    console.log("Header not valid: " + header);
                            }
                        }
                    });
                    return transaction;
                });

            setTransactions(newTransactions);
            setSelectedMethodId(null);
        };

        reader.readAsText(file);
        event.target.value = null;
    };

    const handleUploadClick = () => {
        setIsMethodDialogOpen(true);
    };

    const handleChange = (index, field, value) => {
        const newTransactions = [...transactions];
        newTransactions[index] = {
            ...newTransactions[index],
            [field]: value
        };

        setTransactions(newTransactions);
    };

    const addRow = () => {
        setTransactions([...transactions, {
            amount: '',
            date: new Date().toISOString().split('T')[0],
            name: '',
            method: '',
            budget_category: '',
            reward_category: '',
            comment: '',
            source: 'user'
        }]);
    };

    const removeRow = (index) => {
        setTransactions(transactions.filter((_, i) => i !== index));
    };

    const clearTransactions = () => {
        setTransactions([{
            amount: '',
            date: new Date().toISOString().split('T')[0],
            name: '',
            method: '',
            budget_category: '',
            reward_category: '',
            comment: '',
            source: 'user'
        }]);
    };

    const handleClose = () => {
        clearTransactions();
        onClose();
    };

    const handleSubmit = () => {
        let x = createExpenditures({variables: {expenditures: transactions}});
        x.catch(
            (e) => {
                console.log(e)
                alert("Error: " + e.message)
            }
        )
        handleClose();
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="xl" fullWidth>
            <DialogTitle>Add Transactions</DialogTitle>
            <DialogContent>
                {transactions.map((transaction, index) => (
                    <Stack direction="row" spacing={1} key={index} sx={{mt: 2}}>
                        <TextField
                            name="date"
                            label="Date"
                            type="date"
                            value={transaction.date}
                            onChange={(e) => handleChange(index, 'date', e.target.value)}
                            sx={{width: '12%'}}
                        />
                        <TextField
                            name="amount"
                            label="Amount"
                            type="number"
                            value={transaction.amount}
                            onChange={(e) => handleChange(index, 'amount', e.target.value)}
                            sx={{width: '10%'}}
                        />
                        <TextField
                            name="name"
                            label="Description"
                            value={transaction.name}
                            onChange={(e) => handleChange(index, 'name', e.target.value)}
                            sx={{width: '15%'}}
                        />
                        <FormControl sx={{width: '12%'}}>
                            <Autocomplete
                                size="small"
                                options={paymentMethodsData?.paymentMethods || []}
                                value={paymentMethodsData?.paymentMethods?.find(m => m.id === transaction.method) || null}
                                onChange={(_, value) => handleChange(index, 'method', value?.id)}
                                getOptionLabel={(option) => option?.displayName || ''}
                                renderInput={(params) => (
                                    <TextField
                                        {...params}
                                        label="Method"
                                        sx={{
                                            '& .MuiInputBase-root': {
                                                height: '40px',
                                                padding: '0 9px'
                                            }
                                        }}
                                    />
                                )}
                                isOptionEqualToValue={(option, value) => option?.id === value?.id}
                            />
                        </FormControl>
                        <FormControl sx={{width: '15%'}}>
                            <InputLabel>Budget Category</InputLabel>
                            <Select
                                name="budget_category"
                                value={transaction.budget_category}
                                onChange={(e) => handleChange(index, 'budget_category', e.target.value)}
                                label="Budget Category"
                            >
                                {budget?.expenses?.map((expense) => (
                                    <MenuItem key={expense.category} value={expense.category}>
                                        {expense.category}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                        <FormControl sx={{width: '15%'}}>
                            <Autocomplete
                                options={COMMON_REWARDS_CATEGORIES}
                                value={transaction.reward_category || ''}
                                onChange={(_, newValue) => handleChange(index, 'reward_category', newValue)}
                                renderInput={(params) => (
                                    <TextField
                                        {...params}
                                        label="Reward Category"
                                        sx={{
                                            '& .MuiInputBase-root': {
                                                height: '40px',
                                                padding: '0 9px'
                                            }
                                        }}
                                    />
                                )}
                            />
                        </FormControl>
                        <TextField
                            name="comment"
                            label="Comment"
                            value={transaction.comment}
                            onChange={(e) => handleChange(index, 'comment', e.target.value)}
                            sx={{width: '15%'}}
                        />
                        <IconButton
                            onClick={() => removeRow(index)}
                            disabled={transactions.length === 1}
                        >
                            <DeleteOutlined/>
                        </IconButton>
                    </Stack>
                ))}
            </DialogContent>
            <DialogContent sx={{ pb: 0 }}>
                <Typography variant="subtitle2" sx={{ textAlign: 'right' }}>
                    {transactions.length} transaction{transactions.length !== 1 ? 's' : ''}
                </Typography>
            </DialogContent>
            <DialogActions>
                <input
                    type="file"
                    accept=".csv"
                    ref={fileInputRef}
                    onChange={handleFileUpload}
                    style={{display: 'none'}}
                />
                <PaymentMethodDialog
                    open={isMethodDialogOpen}
                    onClose={() => setIsMethodDialogOpen(false)}
                    onSelect={handleMethodSelect}
                    paymentMethods={paymentMethodsData?.paymentMethods || []}
                />
                <Button
                    startIcon={<UploadOutlined/>}
                    onClick={handleUploadClick}
                    size="small"
                >
                    Upload CSV
                </Button>
                <Button
                    startIcon={<PlusOutlined/>}
                    onClick={addRow}
                >
                    Add Row
                </Button>
                <Button
                    color="error"
                    onClick={clearTransactions}
                >
                    Clear All
                </Button>
                <Stack direction="row" spacing={1}>
                    <Button onClick={handleClose}>Cancel</Button>
                    <Button onClick={handleSubmit} variant="contained">Save</Button>
                </Stack>
            </DialogActions>
        </Dialog>
    );
}