import { gql } from "@apollo/client";
import { useMutation, useQuery } from "@apollo/client/react";

export function useBudgets(count) {
  return useQuery(LIST_BUDGETS, {
    variables: { count },
  });
}

export const useRewardCards = ({ issuer, name, region }) => {
  return useQuery(GET_REWARD_CARDS, {
    variables: {
      issuer, name, region
    },
  });
}
export const usePaymentMethods = () => {
  return useQuery(GET_PAYMENT_METHODS);
};
export function useCreateBudget({ name, incomes, expenses }) {
  return useMutation(CREATE_BUDGET, {
    variables: {
      name,
      incomes,
      expenses,
    },
    refetchQueries: [{ query: LIST_BUDGETS }],
  });
}

export function useUpdateBudget({ id, name, incomes, expenses }) {
  return useMutation(UPDATE_BUDGET, {
    variables: {
      id,
      name,
      incomes,
      expenses,
    },
    refetchQueries: [{ query: GET_BUDGET_BY_ID }],
  });
}

export function useCreateExpenditures(expenditures) {
  return useMutation(UPSERT_EXPENDITURE, {
    variables: {
      expenditures,
    },
    refetchQueries: [{ query: RECENT_EXPENDITURES }],
  });
}

export function useCreatePaymentMethod() {
  return useMutation(CREATE_PAYMENT_METHOD, {
    refetchQueries: [{ query: GET_PAYMENT_METHODS }]
  });
}

export function useDeletePaymentMethod() {
  return useMutation(DELETE_PAYMENT_METHOD, {
    refetchQueries: [{ query: GET_PAYMENT_METHODS }]
  });
}

export const GET_BUDGET_BY_ID = gql`
  query Budget($id: ID!) {
    budget(id: $id) {
      id
      owner
      name
      incomes {
        source
        amount
      }
      expenses {
        id
        category
        amount
        isFixed
        isSlack
      }
    }
  }
`;

export const LIST_BUDGETS = gql`
  query Budgets($count: Int) {
    budgets(first: $count) {
      id
      owner
      name
      incomes {
        source
        amount
      }
      expenses {
        id
        category
        amount
        isFixed
        isSlack
      }
    }
  }
`;

export const RECENT_EXPENDITURES = gql`
  query Expenditures($filter: String, $since: String, $until: String, $paymentMethod: String, $category: String $count: Int, $offset: Int) {
    expenditures(filter: $filter, since: $since, until: $until, paymentMethod: $paymentMethod, category: $category, count: $count, offset: $offset) {
      id
      owner
      name
      amount
      date
      method
      budget_category
      reward_category
      comment
      source
    }
  }
`;

export const AGGREGATE_EXPENDITURES = gql`
  query AggregatedExpenditures(
    $since: String
    $until: String
    $span: Timespan
    $groupBy: GroupBy
  ) {
    aggregatedExpenditures(
      since: $since
      until: $until
      span: $span
      groupBy: $groupBy
      aggregation: SUM
    ) {
      groupByCategory
      amount
      spanStart
      span
    }
  }
`;

export const GET_REWARD_CARDS = gql`
  query GetRewardCards($issuer: String, $name: String, $region: String) {
    rewardCards(issuer: $issuer, name: $name, region: $region) {
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
`;

export const GET_PAYMENT_METHODS = gql`
  query GetPaymentMethods {
    paymentMethods {
      id
      displayName
      acquiredDate
      cancelByDate
      cardType
      rewards {
        name
        issuer
        rewardType
        categories {
          category
          rate
        }
      }
    }
  }
`;

export const CREATE_BUDGET = gql`
  mutation CreateBudget(
    $name: String!
    $incomes: [IncomeInput]
    $expenses: [ExpenseInput]
  ) {
    createBudget(
      input: { name: $name, incomes: $incomes, expenses: $expenses }
    ) {
      id
      name
      incomes {
        source
        amount
      }
      expenses {
        id
        category
        amount
        isFixed
        isSlack
      }
    }
  }
`;

export const UPDATE_BUDGET = gql`
  mutation UpdateBudget(
    $id: ID!
    $name: String
    $incomes: [IncomeInput]
    $expenses: [ExpenseInput]
  ) {
    updateBudget(
      input: { id: $id, name: $name, incomes: $incomes, expenses: $expenses }
    ) {
      id
      name
      incomes {
        source
        amount
      }
      expenses {
        id
        category
        amount
        isFixed
        isSlack
      }
    }
  }
`;

export const UPSERT_EXPENDITURE = gql`
  mutation CreateExpenditures($expenditures: [ExpenditureInput]!) {
    createExpenditures(input: $expenditures)
  }
`

export const CREATE_PAYMENT_METHOD = gql`
  mutation CreatePaymentMethod($input: PaymentMethodInput!) {
    createPaymentMethod(input: $input) {
      id
      displayName
      acquiredDate
      cancelByDate
      cardType
    }
  }
`;

export const UPDATE_PAYMENT_METHOD = gql`
  mutation UpdatePaymentMethod($id: ID!, $input: PaymentMethodInput!) {
    updatePaymentMethod(id: $id, input: $input) {
      id
    }
  }
`;

export const DELETE_PAYMENT_METHOD = gql`
  mutation DeletePaymentMethod($id: ID!) {
    deletePaymentMethod(id: $id)
  }
`;

export const GET_PAYMENT_METHOD_NAMES = gql`
  query GetPaymentMethodNames {
    paymentMethods {
      id
      displayName
    }
  }
`;

export const usePaymentMethodNames = () => {
  return useQuery(GET_PAYMENT_METHOD_NAMES);
};