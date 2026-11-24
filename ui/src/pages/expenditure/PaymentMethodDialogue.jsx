import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';

export const PaymentMethodDialog = ({ open, onClose, onSelect, paymentMethods }) => {
    return (
        <Dialog open={open} onClose={() => onClose()}>
            <DialogTitle>Select Payment Method</DialogTitle>
            <DialogContent>
                <FormControl fullWidth sx={{ mt: 2 }}>
                    <Select
                        value=""
                        onChange={(e) => onSelect(e.target.value)}
                    >
                        {paymentMethods?.map((method) => (
                            <MenuItem key={method.id} value={method.id}>
                                {method.displayName}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>
            </DialogContent>
        </Dialog>
    );
};