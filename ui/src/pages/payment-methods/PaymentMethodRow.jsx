import { useState } from 'react';
import { Button, Stack, Typography } from '@mui/material';
import MainCard from 'components/MainCard';
import { useDeletePaymentMethod } from "../../api/graph";
import { DeleteOutlined } from '@ant-design/icons';
import EditOutlined from "@ant-design/icons/EditOutlined";

import EditPaymentMethodDialog from "./EditPaymentMethodDialogue";

const PaymentMethodRow = ({ method }) => {
    const [deletePaymentMethod] = useDeletePaymentMethod();
    const [editOpen, setEditOpen] = useState(false);
    const handleDelete = async () => {
        try {
            await deletePaymentMethod({
                variables: { id: method.id }
            });
        } catch (error) {
            console.error('Failed to delete payment method:', error);
        }
    };

    return (
        <MainCard sx={{ mb: 2 }}>
            <Stack spacing={2}>
                <Stack direction="row" alignItems="center" justifyContent="space-between">
                    <Stack spacing={1}>
                        <Typography variant="h6">{method.displayName}</Typography>
                        <Typography variant="body2" color="text.secondary">
                            Added on {new Date(method.acquiredDate).toLocaleDateString()}
                            {method.cancelByDate && ` · Cancel by ${new Date(method.cancelByDate).toLocaleDateString()}`}
                        </Typography>
                    </Stack>
                    <Stack>
                    <Button
                        variant="outlined"
                        color="primary"
                        onClick={() => setEditOpen(true)}
                        startIcon={<EditOutlined />}
                    >
                        Edit
                    </Button>
                    <Button
                        variant="outlined"
                        color="error"
                        onClick={handleDelete}
                        startIcon={<DeleteOutlined />}
                    >
                        Delete
                    </Button>
                    </Stack>
                </Stack>

                {method.rewards && (
                    <Stack spacing={1}>
                        <Typography variant="subtitle1">
                            {method.rewards.issuer} {method.rewards.name} · {method.rewards.rewardType}
                        </Typography>
                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} flexWrap="wrap">
                            {method.rewards.categories?.map((cat, index) => (
                                <Typography
                                    key={index}
                                    variant="body2"
                                    sx={{
                                        bgcolor: 'action.hover',
                                        px: 1,
                                        py: 0.5,
                                        borderRadius: 1,
                                        display: 'inline-flex'
                                    }}
                                >
                                    {cat.category}: {cat.rate}X
                                </Typography>
                            ))}
                        </Stack>
                    </Stack>
                )}
            </Stack>
            <EditPaymentMethodDialog
                open={editOpen}
                onClose={() => setEditOpen(false)}
                method={method}
            />
        </MainCard>
    );
};

export default PaymentMethodRow;