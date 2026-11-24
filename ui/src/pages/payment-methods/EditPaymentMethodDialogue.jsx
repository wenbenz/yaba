import { Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField, Stack } from '@mui/material';
import { useState, useEffect } from 'react';
import {BrowseFilters} from "../browse";
import Box from "@mui/material/Box";
import {UPDATE_PAYMENT_METHOD, GET_PAYMENT_METHODS, useRewardCards} from "../../api/graph";
import Typography from "@mui/material/Typography";
import SearchOutlined from "@ant-design/icons/SearchOutlined";
import {useMutation} from "@apollo/client/react";

export default function EditPaymentMethodDialog({ open, onClose, method }) {
    const NULL_UUID = "00000000-0000-0000-0000-000000000000";

    const [formData, setFormData] = useState({
        displayName: '',
        acquiredDate: null,
        cancelByDate: null,
        cardType: '',
    });
    const [filters, setFilters] = useState({
        issuer: '',
        name: '',
        region: '',
    });
    const [selectedCard, setSelectedCard] = useState(null);
    const [showBrowser, setShowBrowser] = useState(false);

    const {loading, error, data} = useRewardCards(filters)
    const [updatePaymentMethod] = useMutation(UPDATE_PAYMENT_METHOD, {
        refetchQueries: [{ query: GET_PAYMENT_METHODS }]
    });

    const handleFilterChange = (event) => {
        const {name, value} = event.target;
        setFilters((prev) => ({
            ...prev,
            [name]: value || '',
        }));
    };

    useEffect(() => {
        if (method) {
            setFormData({
                id: method.id,
                displayName: method.displayName,
                acquiredDate: method.acquiredDate || null,
                cancelByDate: method.cancelByDate || null,
                cardType: method.cardType === NULL_UUID ? null : method.cardType || '',
            });
            setSelectedCard(method.rewards)
        }
    }, [method]);

    const handleSubmit = (e) => {
        e.preventDefault();
        updatePaymentMethod({
            variables: {
                id: method.id,
                input: {
                    displayName: formData.displayName,
                    acquiredDate: formData.acquiredDate,
                    cancelByDate: formData.cancelByDate,
                    cardType: formData.cardType || NULL_UUID
                }
            }
        });
        onClose();
    };

    const formatDate = (date) => {
        if (!date) return '';
        return new Date(date).toISOString().split('T')[0];
    };

    const handleDateChange = (field) => (e) => {
        const value = e.target.value || null;
        setFormData(prev => ({ ...prev, [field]: value }));
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>Edit Payment Method</DialogTitle>
            <form onSubmit={handleSubmit}>
                <DialogContent>
                    <Stack spacing={2}>
                        <TextField
                            label="Display Name"
                            value={formData.displayName}
                            onChange={(e) => setFormData(prev => ({ ...prev, displayName: e.target.value }))}
                            fullWidth
                            required
                        />
                        <Stack direction="row" spacing={2}>
                            <TextField
                                label="Acquired Date"
                                type="date"
                                value={formatDate(formData.acquiredDate)}
                                onChange={handleDateChange('acquiredDate')}
                                fullWidth
                                InputLabelProps={{ shrink: true }}
                            />
                            <TextField
                                label="Cancel By Date"
                                type="date"
                                value={formatDate(formData.cancelByDate)}
                                onChange={handleDateChange('cancelByDate')}
                                fullWidth
                                InputLabelProps={{ shrink: true }}
                            />
                        </Stack>
                        <Stack
                            direction="row"
                            spacing={2}
                            alignItems="center"
                            >
                            <Typography variant={"subtitle1"}>Card Type:</Typography>
                            <Typography variant={"subtitle2"}>{selectedCard?.issuer} {selectedCard?.name}</Typography>
                            <Button
                                variant="outlined"
                                color="primary"
                                onClick={() => {
                                    setShowBrowser(!showBrowser);
                                    setSelectedCard(null);
                                }}>
                                <SearchOutlined />
                            </Button>
                        </Stack>
                        <Box sx={{ display: showBrowser ? 'block' : 'none' }}>
                            <BrowseFilters
                                filters={filters}
                                onFilterChange={handleFilterChange}
                            />
                            {(loading && <p>Loading...</p>)
                                || (error && <p>Error: {error.message}</p>)
                                || data?.rewardCards?.map((card) => (
                                    <Box key={card.id} sx={{ padding: 2, border: '1px solid #ccc', marginTop: 2 }}>
                                        <Stack direction="row" justifyContent="space-between" alignItems="center">
                                            <h4>{card.issuer} {card.name}</h4>
                                            <Button
                                                variant="outlined"
                                                color="primary"
                                                onClick={() => {
                                                    setFormData(prev => ({...prev, cardType: card.id}));
                                                    setSelectedCard(card);
                                                    setShowBrowser(false);
                                                }}
                                            >Select</Button>
                                        </Stack>
                                    </Box>
                                ))}
                        </Box>
                    </Stack>
                </DialogContent>
                <DialogActions>
                    <Button onClick={onClose}>Cancel</Button>
                    <Button type="submit" variant="contained">Save Changes</Button>
                </DialogActions>
            </form>
        </Dialog>
    );
}