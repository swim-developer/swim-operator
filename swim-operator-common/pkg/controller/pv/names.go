package pv

import "fmt"

func MariaDBServiceName(crName string) string {
	return fmt.Sprintf("%s-mariadb", crName)
}
