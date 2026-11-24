import {useState} from 'react';
import {gql} from '@apollo/client';
import {useMutation} from '@apollo/client/react';
import {Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Stack, TextField,} from '@mui/material';
import {DeleteOutlined} from '@ant-design/icons';
import {Autocomplete} from "@mui/material";
import {COMMON_ISSUERS, COMMON_REGIONS, COMMON_REWARD_TYPES, COMMON_REWARDS_CATEGORIES} from "../../utils/constants";
import Typography from "@mui/material/Typography";

const CREATE_REWARD_CARD = gql`
    mutation CreateRewardCard($input: RewardCardInput!) {
        createRewardCard(input: $input) {
            id
            name
            issuer
            region
            version
            rewardType
            categories {
                category
                rate
            }
        }
    }
`

const CategoryInput = ({ value, onChange, error }) => {
    return (
        <Stack direction="row" spacing={2} alignItems="center" sx={{ flex: 1 }}>
            <Autocomplete
                freeSolo
                options={COMMON_REWARDS_CATEGORIES}
                value={value}
                onChange={(_, newValue) => onChange(newValue)}
                sx={{ flex: 1 }}
                renderInput={(params) => (
                    <TextField
                        {...params}
                        required
                        label="Category"
                        error={error}
                        onChange={(e) => onChange(e.target.value)}
                        sx={{
                            '& .MuiInputBase-root': {
                                height: '40px',
                                padding: '0 9px'
                            }
                        }}
                    />
                )}
            />
        </Stack>
    );
};

const RewardTypeInput = ({ value, onChange }) => {
    return (
        <Autocomplete
            freeSolo
            options={COMMON_REWARD_TYPES}
            value={value}
            onChange={(_, newValue) => onChange({ target: { name: 'rewardType', value: newValue }})}
            renderInput={(params) => (
                <TextField
                    {...params}
                    required
                    label="Reward Type"
                    onChange={(e) => onChange({ target: { name: 'rewardType', value: e.target.value }})}
                    sx={{
                        '& .MuiInputBase-root': {
                            height: '40px',
                            padding: '0 9px'
                        }
                    }}
                />
            )}
        />
    );
};

const IssuerInput = ({ value, onChange }) => {
    return (
        <Autocomplete
            freeSolo
            options={COMMON_ISSUERS}
            value={value}
            onChange={(_, newValue) => onChange({ target: { name: 'issuer', value: newValue }})}
            renderInput={(params) => (
                <TextField
                    {...params}
                    required
                    label="Issuer"
                    onChange={(e) => onChange({ target: { name: 'issuer', value: e.target.value }})}
                    sx={{
                        '& .MuiInputBase-root': {
                            height: '40px',
                            padding: '0 9px'
                        }
                    }}
                />
            )}
        />
    );
};

const RegionInput = ({ value, onChange }) => {
    return (
        <Autocomplete
            options={COMMON_REGIONS}
            value={value}
            onChange={(_, newValue) => onChange({ target: { name: 'region', value: newValue }})}
            renderInput={(params) => (
                <TextField
                    {...params}
                    required
                    label="Region"
                    sx={{
                        '& .MuiInputBase-root': {
                            height: '40px',
                            padding: '0 9px'
                        }
                    }}
                />
            )}
        />
    );
};
export default function CreateCardDialog({open, onClose, onSuccess, currentFilters}) {
    const [formData, setFormData] = useState({
        name: currentFilters?.name || '',
        issuer: currentFilters?.issuer || '',
        region: currentFilters?.region || '',
        rewardType: '',
        rewardCategories: [{category: '', rate: ''}]
    });

    const [createCard] = useMutation(CREATE_REWARD_CARD, {
        update: (cache, {data: {createRewardCard}}) => {
            cache.modify({
                fields: {
                    rewardCards(existingCards = []) {
                        const newCardRef = cache.writeFragment({
                            data: createRewardCard,
                            fragment: gql`
                                fragment NewCard on RewardCard {
                                    id
                                    name
                                    issuer
                                    region
                                    version
                                    rewardType
                                    categories {
                                        category
                                        rate
                                    }
                                }
                            `
                        });
                        return [...existingCards, newCardRef];
                    }
                }
            });
        }
    });

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const input = {
                ...formData,
                rewardCategories: formData.rewardCategories.map(cat => ({
                    ...cat,
                    rate: parseFloat(cat.rate)
                }))
            };
            await createCard({variables: {input}});
            onSuccess?.();
            onClose();
        } catch (error) {
            console.error('Error creating card:', error);
        }
    };

    const handleChange = (e) => {
        setFormData(prev => ({
            ...prev,
            [e.target.name]: e.target.value
        }));
    };

    const handleCategoryChange = (index, field, value) => {
        setFormData(prev => ({
            ...prev,
            rewardCategories: prev.rewardCategories.map((cat, i) =>
                i === index ? {...cat, [field]: value} : cat
            )
        }));
    };

    const addCategory = () => {
        setFormData(prev => ({
            ...prev,
            rewardCategories: [...prev.rewardCategories, {category: '', rate: ''}]
        }));
    };

    const removeCategory = (index) => {
        setFormData(prev => ({
            ...prev,
            rewardCategories: prev.rewardCategories.filter((_, i) => i !== index)
        }));
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <form onSubmit={handleSubmit}>
                <DialogTitle>Add New Reward Card</DialogTitle>
                <DialogContent>
                    <Stack spacing={2} sx={{mt: 1}}>
                        <TextField
                            required
                            fullWidth
                            name="name"
                            label="Card Name"
                            value={formData.name}
                            onChange={handleChange}
                        />
                        <IssuerInput
                            value={formData.issuer}
                            onChange={handleChange}
                        />
                        <RegionInput
                            value={formData.region}
                            onChange={handleChange}
                        />
                        <RewardTypeInput
                            value={formData.rewardType}
                            onChange={handleChange}
                        />

                        {formData.rewardCategories.map((cat, index) => (
                            <Stack key={index} direction="row" spacing={2} alignItems="center">
                                <CategoryInput
                                    value={cat.category}
                                    onChange={(value) => handleCategoryChange(index, 'category', value)}
                                />
                                <TextField
                                    required
                                    label="Rate"
                                    type="number"
                                    inputProps={{
                                        step: "1",
                                        style: { textAlign: 'right' }
                                    }}                                    value={cat.rate}
                                    onChange={(e) => handleCategoryChange(index, 'rate', e.target.value)}
                                />
                                <Typography sx={{ minWidth: 20 }}>
                                    {formData.rewardType?.toLowerCase() === 'cashback' ? '% cashback' : 'X points'}
                                </Typography>
                                {formData.rewardCategories.length > 1 && (
                                    <IconButton onClick={() => removeCategory(index)} size="small">
                                        <DeleteOutlined/>
                                    </IconButton>
                                )}
                            </Stack>
                        ))}

                        <Button type="button" onClick={addCategory}>
                            Add Category
                        </Button>
                    </Stack>
                </DialogContent>
                <DialogActions>
                    <Button onClick={onClose}>Cancel</Button>
                    <Button type="submit" variant="contained">Create</Button>
                </DialogActions>
            </form>
        </Dialog>
    );
}