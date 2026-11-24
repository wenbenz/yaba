import { Grid, Typography } from '@mui/material';
import MainCard from "../../components/MainCard";
import PaymentMethodsList from "./PaymentMethodsList";

const PaymentMethods = () => {
    return (
        <Grid container rowSpacing={4.5} columnSpacing={2.75}>
            <Grid item xs={12} sx={{ mb: -2.25 }}>
                <Typography variant="h5">Payment Methods</Typography>
            </Grid>

            <Grid item xs={12}>
                <MainCard>
                    <PaymentMethodsList />
                </MainCard>
            </Grid>
        </Grid>
    );
};

export default PaymentMethods;