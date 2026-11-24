import { Alert } from "@mui/material";
import {ExclamationCircleFilled} from "@ant-design/icons";
import { useCreateBudget } from "../../api/graph";
import { useEffect } from "react";

export default function CreateBudget({ budget }) {
  const [createBudget] = useCreateBudget(budget);

  useEffect(() => {
    createBudget(budget);
  }, []);

  return (
    <Alert color="primary" icon={<ExclamationCircleFilled />}>
      No existing budget found. New budget created from template.
    </Alert>
  );
}
