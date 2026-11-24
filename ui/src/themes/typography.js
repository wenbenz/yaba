// ==============================|| DEFAULT THEME - TYPOGRAPHY ||============================== //

export default function Typography(fontFamily) {
  return {
    htmlFontSize: 16,
    fontFamily,
    fontWeightLight: 300,
    fontWeightRegular: 400,
    fontWeightMedium: 500,
    fontWeightBold: 600,
    h1: {
      fontWeight: 500,
      fontSize: "2rem",
      lineHeight: 1.25,
      letterSpacing: "-0.02em",
    },
    h2: {
      fontWeight: 500,
      fontSize: "1.625rem",
      lineHeight: 1.3,
      letterSpacing: "-0.015em",
    },
    h3: {
      fontWeight: 500,
      fontSize: "1.375rem",
      lineHeight: 1.35,
      letterSpacing: "-0.01em",
    },
    h4: {
      fontWeight: 500,
      fontSize: "1.125rem",
      lineHeight: 1.4,
    },
    h5: {
      fontWeight: 500,
      fontSize: "0.9375rem",
      lineHeight: 1.5,
    },
    h6: {
      fontWeight: 400,
      fontSize: "0.875rem",
      lineHeight: 1.57,
    },
    caption: {
      fontWeight: 400,
      fontSize: "0.75rem",
      lineHeight: 1.66,
    },
    body1: {
      fontSize: "0.875rem",
      lineHeight: 1.6,
    },
    body2: {
      fontSize: "0.75rem",
      lineHeight: 1.66,
    },
    subtitle1: {
      fontSize: "0.875rem",
      fontWeight: 500,
      lineHeight: 1.57,
    },
    subtitle2: {
      fontSize: "0.75rem",
      fontWeight: 500,
      lineHeight: 1.66,
    },
    overline: {
      lineHeight: 1.66,
    },
    button: {
      textTransform: "capitalize",
    },
  };
}
