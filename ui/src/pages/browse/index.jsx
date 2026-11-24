import {useState} from 'react';
import {
    Box,
    Button,
    Card,
    CardContent,
    CircularProgress,
    Grid,
    Stack,
    TextField,
    Typography
} from '@mui/material';
import CreateCardDialog from './CreateCardDialogue';
import {COMMON_ISSUERS, COMMON_REGIONS} from '../../utils/constants';
import {PlusCircleFilled} from '@ant-design/icons';
import AddPaymentMethodDialog from "./AddPaymentMethodDialogue";
import {useRewardCards} from "../../api/graph";
import {Autocomplete} from "@mui/material";

export const BrowseFilters = ({ filters, onFilterChange }) => {
    return (
        <Grid container spacing={2}>
            <Grid item xs={12} sm={4}>
                <Autocomplete
                    options={COMMON_REGIONS}
                    value={filters.region}
                    onChange={(_, value) => onFilterChange({
                        target: { name: 'region', value }
                    })}
                    renderInput={(params) => (
                        <TextField
                            {...params}
                            fullWidth
                            name="region"
                            label="Region"
                        />
                    )}
                />
            </Grid>
            <Grid item xs={12} sm={4}>
                <Autocomplete
                    freeSolo
                    options={COMMON_ISSUERS}
                    value={filters.issuer}
                    onChange={(_, value) => onFilterChange({
                        target: { name: 'issuer', value }
                    })}
                    renderInput={(params) => (
                        <TextField
                            {...params}
                            fullWidth
                            name="issuer"
                            label="Issuer"
                        />
                    )}
                />
            </Grid>
            <Grid item xs={12} sm={4}>
                <Autocomplete
                    freeSolo
                    options={[]} // Empty array since we don't need suggestions
                    value={filters.name}
                    onInputChange={(_, value) => onFilterChange({
                        target: { name: 'name', value }
                    })}
                    renderInput={(params) => (
                        <TextField
                            {...params}
                            fullWidth
                            name="name"
                            label="Card Name"
                        />
                    )}
                />
            </Grid>
        </Grid>
    );
};

export default function BrowseCards() {
    const [filters, setFilters] = useState({
        issuer: '',
        name: '',
        region: '',
    });
    const [selectedCard, setSelectedCard] = useState(null);
    const [openCreateCardDialogue, setOpenCreateCardDialogue] = useState(false);
    const [openAddPaymentMethodDialogue, setOpenAddPaymentMethodDialogue] = useState(false);

    const handleOpenAddPaymentMethodDialog = (card) => {
        setSelectedCard(card);
        setOpenAddPaymentMethodDialogue(true);
    };

    const handleCloseAddPaymentMethodDialog = () => {
        setOpenAddPaymentMethodDialogue(false);
        setSelectedCard(null);
    };

    const {loading, error, data} = useRewardCards(filters)

    const handleFilterChange = (event) => {
        const {name, value} = event.target;
        setFilters((prev) => ({
            ...prev,
            [name]: value || '',
        }));
    };

    return (
        <Grid container maxWidth="xl">
            <Grid item xs={12} sx={{mb: 3}}>
                <Typography variant="h5" sx={{ mb: 3 }}>
                    Browse Cards
                </Typography>
                <BrowseFilters
                    filters={filters}
                    onFilterChange={handleFilterChange}
                />
            </Grid>
            {loading && (
                <Box sx={{display: 'flex', justifyContent: 'center', my: 4}}>
                    <CircularProgress/>
                </Box>
            )}

            {error && (
                <Typography color="error" sx={{my: 2}}>
                    Error: {error.message}
                </Typography>
            )}

            {!loading && !error && (
                <Grid container spacing={3}>
                    {data?.rewardCards?.map((card) => (
                        <Grid item xs={12} sm={6} md={4} key={card.id}>
                            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
                                <CardContent sx={{
                                    height: '100%',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    justifyContent: 'space-between'
                                }}>
                                    <Stack spacing={1}>
                                        <Stack direction="row" justifyContent="space-between" alignItems="center">
                                            <Typography variant="h6">
                                                {card.issuer} {card.name}
                                            </Typography>
                                            <Typography variant="body1">
                                                Rewards: {card.rewardType}
                                            </Typography>
                                        </Stack>
                                        <Box sx={{
                                            display: { md: 'flex' },
                                            flexWrap: 'wrap',
                                            gap: 1
                                        }}>
                                            {card.categories?.map((category, index) => (
                                                <Typography
                                                    key={index}
                                                    variant="body2"
                                                    sx={{
                                                        '&:not(:last-child):after': {
                                                            content: {
                                                                md: '"•"'
                                                            },
                                                            mx: 1,
                                                            display: { xs: 'none', md: 'inline' }
                                                        }
                                                    }}
                                                >
                                                    {category.category}: {category.rate}X
                                                </Typography>
                                            ))}
                                        </Box>
                                    </Stack>
                                    <Typography variant="caption" color="text.secondary" sx={{ mt: 2 }}>
                                        Region: {card.region} | Version: {card.version}
                                    </Typography>
                                    <Stack direction="row" justifyContent="flex-end" sx={{ mt: 2 }}>
                                        <Button
                                            variant="contained"
                                            color="primary"
                                            onClick={() => handleOpenAddPaymentMethodDialog(card)}
                                        >
                                            Add to Payment Methods
                                        </Button>
                                    </Stack>
                                </CardContent>
                            </Card>
                        </Grid>
                    ))}

                </Grid>
            )}

            <Box sx={{position: 'fixed', bottom: 24, right: 24}}>
                <Button
                    variant="contained"
                    color="primary"
                    onClick={() => setOpenCreateCardDialogue(true)}
                >
                    <Stack direction="row" spacing={1} alignItems="center">
                        <PlusCircleFilled/>
                        <Typography>Add New Card</Typography>
                    </Stack>
                </Button>
            </Box>

            <CreateCardDialog
                open={openCreateCardDialogue}
                onClose={() => setOpenCreateCardDialogue(false)}
                currentFilters={filters}
                onSuccess={() => {
                    // Optionally update filters to match the new card
                }}
            />

            <AddPaymentMethodDialog
                open={openAddPaymentMethodDialogue}
                onClose={handleCloseAddPaymentMethodDialog}
                card={selectedCard}
            />
        </Grid>
    );
}