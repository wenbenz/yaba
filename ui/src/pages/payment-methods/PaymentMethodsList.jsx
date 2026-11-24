import { Grid } from '@mui/material';
import { usePaymentMethods } from 'api/graph';
import PaymentMethodRow from './PaymentMethodRow';

const PaymentMethodsList = () => {
    const { data } = usePaymentMethods();
    return (
        <Grid container>
            <Grid item xs={12}>
                {data?.paymentMethods.map((method) => (
                    <PaymentMethodRow
                        key={method.id}
                        method={method}
                    />
                ))}
            </Grid>
        </Grid>
    );
};

export default PaymentMethodsList;