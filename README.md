# Yet Another Budgeting App (YABA)
![BuildStatus](https://img.shields.io/github/actions/workflow/status/wenbenz/yaba/build.yml)
[![codecov](https://codecov.io/gh/wenbenz/yaba/graph/badge.svg?token=SXZB119QHO)](https://codecov.io/gh/wenbenz/yaba)

Tracking spending is a pain. Most budgeting tools require manual input of each
transaction, and budget labels don't always line up with credit card labels.
Financial information is also deeply personal, and with data concerns these
days, it's difficult to trust that the data will be stored with the level of
care that is warranted. This is why YABA was created. YABA intends to solve
these problems by:
- offering a budgeting tool that can be self-hosted
- creating features that enable users to define speding categories and track
spending over time
- display the credit card reward categories of actual spending to find the best
credit card for the budget
- reduce manual input by adding PFD imports and automatic assignments for
similar transactions.

## Quick Start
Create a password file.
```shell
echo "POSTGRES PASSWORD HERE" > db_password.txt
```

Pull docker image.
```shell
docker pull wenbenz/yaba
```

Choose a docker-compose from `config/examples` and run it with
```shell
docker compose up
```
