import { Link as RouterLink } from "react-router-dom";

import Container from "@mui/material/Container";
import Link from "@mui/material/Link";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";

export default function AuthFooter() {
  return (
    <Container maxWidth="xl">
      <Stack direction="row" justifyContent="center" spacing={2} alignItems="center">
        <Typography variant="subtitle2" color="secondary">
          © {new Date().getFullYear()} Yaba
        </Typography>
        <Typography variant="subtitle2" color="secondary">·</Typography>
        <Link component={RouterLink} to="/terms" variant="subtitle2" color="secondary">
          Terms of Service
        </Link>
        <Typography variant="subtitle2" color="secondary">·</Typography>
        <Link component={RouterLink} to="/privacy" variant="subtitle2" color="secondary">
          Privacy Policy
        </Link>
      </Stack>
    </Container>
  );
}
