// material-ui
import Link from "@mui/material/Link";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableRow from "@mui/material/TableRow";
import Box from "@mui/material/Box";

// third-party
import { NumericFormat } from "react-number-format";
import Loader from "../../components/Loader";
import {useMemo, useState} from "react";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import {useQuery} from "@apollo/client/react";
import {RECENT_EXPENDITURES, usePaymentMethodNames} from "../../api/graph";
import Divider from "@mui/material/Divider";
import {RecentTransactionsTableHeader} from "./TableHeaders";

export default function RecentTransactionsTable() {
  const [offset, setOffset] = useState(0);
  const [expenditures, setExpenditures] = useState([]);
  const { data: paymentMethodsData } = usePaymentMethodNames();
  const [queryVars, setQueryVars] = useState({
    count: 10,
    offset: 0,
    since: null,
    until: null,
    paymentMethod: null,
    category: null,
  });

  const handleFilterChange = (filters) => {
    setOffset(0);
    setQueryVars((prev) => {
      let newFilters = {
        ...prev,
        ...Object.fromEntries(
            Object.entries(filters)
                .filter(([_, value]) => value !== null)
        ),
        offset: 0,
      }
        return newFilters;
    });
  };

  const handleLoadMore = () => {
    const newOffset = offset + 10;
    setOffset(newOffset);
    setQueryVars((prev) => ({
      ...prev,
      offset: newOffset,
    }));
  };

  const { loading, error, data } = useQuery(RECENT_EXPENDITURES, {
    variables: queryVars,
  });

  useMemo(() => {
    if (data) {
      if (offset === 0) {
        setExpenditures(data.expenditures);
      } else {
        setExpenditures((prev) => [...prev, ...data.expenditures]);
      }
    }
  }, [data, offset]);

  if (loading) return <Loader />;
  if (error) return <p>Error : {error.message}</p>;

  return (
    <Box>
      <TableContainer
        sx={{
          width: "100%",
          overflowX: "auto",
          position: "relative",
          display: "block",
          maxWidth: "100%",
          "& td, & th": { whiteSpace: "nowrap" },
        }}
      >
        <Table aria-labelledby="tableTitle">
          <RecentTransactionsTableHeader filters={queryVars} onFilterChange={handleFilterChange} />
          <TableBody>
            {expenditures.map(
                ({
                  id,
                  date,
                  name,
                  amount,
                  method,
                  budget_category,
                  reward_category,
                  comment,
                }) => {
                  const labelId = `enhanced-table-checkbox-${id}`;

                  return (
                    <TableRow
                      hover
                      role="checkbox"
                      sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
                      tabIndex={-1}
                      key={id}
                    >
                      <TableCell component="th" id={labelId} scope="row">
                        <Link color="secondary"> {date}</Link>
                      </TableCell>
                      <TableCell>{name}</TableCell>
                      <TableCell>
                        <NumericFormat
                          value={amount}
                          displayType="text"
                          thousandSeparator
                          prefix="$"
                        />
                      </TableCell>
                      <TableCell>{paymentMethodsData?.paymentMethods?.find(m => m.id === method)?.displayName}</TableCell>
                      <TableCell>{budget_category}</TableCell>
                      <TableCell>{reward_category}</TableCell>
                      <TableCell align="right">{comment}</TableCell>
                    </TableRow>
                  );
                },
              )}
          </TableBody>
        </Table>
      </TableContainer>
      <Stack>
        <Divider />
        <Button variant="text" onClick={handleLoadMore}>
          <Typography variant={"subtitle1"}>Load More</Typography>
        </Button>
      </Stack>
    </Box>
  );
}

// RecentTransactionsTableHeader.propTypes = { order: PropTypes.any, orderBy: PropTypes.string };
