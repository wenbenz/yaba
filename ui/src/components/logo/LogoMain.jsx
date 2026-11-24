import { useTheme } from "@mui/material/styles";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

const Logo = () => {
  const theme = useTheme();

  return (
    <Box sx={{ display: "flex", alignItems: "center", gap: 1.25 }}>
      <svg
        width="30"
        height="30"
        viewBox="0 0 30 30"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
      >
        <rect width="30" height="30" rx="8" fill={theme.palette.primary.main} />
        <path
          d="M9 8L15 15.5L21 8M15 15.5V22"
          stroke="white"
          strokeWidth="2.75"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
      <Typography
        sx={{
          fontWeight: 600,
          fontSize: "1.1rem",
          letterSpacing: "-0.02em",
          color: theme.palette.primary.dark,
          lineHeight: 1,
          userSelect: "none",
        }}
      >
        yaba
      </Typography>
    </Box>
  );
};

export default Logo;
