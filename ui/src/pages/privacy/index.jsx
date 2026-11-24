import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import Divider from "@mui/material/Divider";
import Typography from "@mui/material/Typography";
import { Link as RouterLink } from "react-router-dom";
import Link from "@mui/material/Link";

import Logo from "components/logo";

export default function PrivacyPolicy() {
  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default", py: 6 }}>
      <Container maxWidth="md">
        <Box sx={{ mb: 4 }}>
          <Logo />
        </Box>

        <Typography variant="h2" gutterBottom>
          Privacy Policy
        </Typography>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          Last updated: April 2026
        </Typography>

        <Divider sx={{ my: 3 }} />

        <Typography variant="h5" gutterBottom>
          1. Self-Hosted by Design
        </Typography>
        <Typography variant="body1" paragraph>
          Yaba runs entirely on your own server. Your financial data never
          leaves your infrastructure. We have no servers, no cloud storage, and
          no access to anything you enter into Yaba.
        </Typography>

        <Typography variant="h5" gutterBottom>
          2. No Data Sharing
        </Typography>
        <Typography variant="body1" paragraph>
          We do not collect, sell, rent, or share your personal information or
          financial data with any third party — ever. There are no analytics,
          no telemetry, and no tracking of any kind built into Yaba.
        </Typography>

        <Typography variant="h5" gutterBottom>
          3. Data You Enter
        </Typography>
        <Typography variant="body1" paragraph>
          All data you enter — budgets, transactions, payment methods — is
          stored in your own database. You own your data entirely. Deleting
          your instance permanently removes all of it.
        </Typography>

        <Typography variant="h5" gutterBottom>
          4. Authentication
        </Typography>
        <Typography variant="body1" paragraph>
          Credentials are stored in your own database and are never transmitted
          to any external service. Use a strong password and secure your server
          accordingly.
        </Typography>

        <Typography variant="h5" gutterBottom>
          5. Changes to This Policy
        </Typography>
        <Typography variant="body1" paragraph>
          If this policy changes, the updated version will be available at this
          URL. We will not make changes that affect how your data is handled
          without updating this document.
        </Typography>

        <Divider sx={{ my: 3 }} />

        <Typography variant="body2" color="text.secondary">
          Also read our{" "}
          <Link component={RouterLink} to="/terms">
            Terms of Service
          </Link>
          .
        </Typography>
      </Container>
    </Box>
  );
}
