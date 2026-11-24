import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    Stack,
    TextField,
    FormControlLabel
} from '@mui/material';
import { useCreatePaymentMethod } from 'api/graph';
import {useEffect, useState} from 'react';
import Checkbox from "@mui/material/Checkbox";

const AddPaymentMethodDialog = ({ open, onClose, card }) => {
    const [createPaymentMethod] = useCreatePaymentMethod();
    const [hasCancelDate, setHasCancelDate] = useState(false);

    const getDefaultCancelDate = (acquiredDate) => {
        const date = new Date(acquiredDate);
        date.setFullYear(date.getFullYear() + 1);
        return date.toISOString().split('T')[0];
    };

    const [formData, setFormData] = useState({
        displayName: '',
        acquiredDate: new Date().toISOString().split('T')[0],
        cancelByDate: getDefaultCancelDate(new Date()),
    });

    useEffect(() => {
        if (open && card) {
            const today = new Date().toISOString().split('T')[0];
            setFormData({
                displayName: `${card.issuer} ${card.name}`,
                acquiredDate: today,
                cancelByDate: getDefaultCancelDate(today),
            });
            setHasCancelDate(false);
        }
    }, [open, card]);

    const handleAcquiredDateChange = (e) => {
        const newAcquiredDate = e.target.value;
        setFormData({
            ...formData,
            acquiredDate: newAcquiredDate,
            cancelByDate: getDefaultCancelDate(newAcquiredDate)
        });
    };

    const handleSubmit = async () => {
        try {
            await createPaymentMethod({
                variables: {
                    input: {
                        ...formData,
                        cancelByDate: hasCancelDate ? formData.cancelByDate : null,
                        cardType: card.id
                    }
                }
            });
            onClose();
        } catch (error) {
            console.error('Failed to create payment method:', error);
        }
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle sx={{ pb: 2 }}>Add Payment Method</DialogTitle>
            <DialogContent sx={{ pb: 3 }}>
                <Stack spacing={3} sx={{ mt: 2 }}>
                    <TextField
                        label="Display Name"
                        value={formData.displayName}
                        onChange={(e) => setFormData({ ...formData, displayName: e.target.value })}
                        fullWidth
                    />
                    <TextField
                        label="Acquired Date"
                        type="date"
                        value={formData.acquiredDate}
                        onChange={handleAcquiredDateChange}
                        fullWidth
                        InputLabelProps={{ shrink: true }}
                    />
                    <FormControlLabel
                        control={
                            <Checkbox
                                checked={hasCancelDate}
                                onChange={(e) => setHasCancelDate(e.target.checked)}
                            />
                        }
                        label="Should this be cancelled by a certain date?"
                    />
                    {hasCancelDate && (
                        <TextField
                            label="Cancel By Date"
                            type="date"
                            value={formData.cancelByDate}
                            onChange={(e) => setFormData({ ...formData, cancelByDate: e.target.value })}
                            fullWidth
                            InputLabelProps={{ shrink: true }}
                        />
                    )}
                </Stack>
            </DialogContent>
            <DialogActions sx={{ px: 3, pb: 3 }}>
                <Button onClick={onClose} variant="outlined">Cancel</Button>
                <Button onClick={handleSubmit} variant="contained">Add Payment Method</Button>
            </DialogActions>
        </Dialog>
    );
};

export default AddPaymentMethodDialog;