package install_bootstrap

import (
	"encoding/json"
	"os"

	"cnb.cool/mliev/dwz/dwz-server/app/dao"
	"cnb.cool/mliev/dwz/dwz-server/app/model"
	"cnb.cool/mliev/dwz/dwz-server/pkg/helper"
)

// AdminFile is where the install endpoint drops the admin credentials so the
// next process boot can create the user once the schema is in place.
const AdminFile = "./config/install_admin.json"

// AdminPayload is the persisted admin user spec.
type AdminPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// Write persists the admin payload to disk.
func Write(admin AdminPayload) error {
	data, err := json.Marshal(admin)
	if err != nil {
		return err
	}
	return os.WriteFile(AdminFile, data, 0600)
}

// Consume reads the bootstrap file (if present), creates the admin user via
// the user DAO, then removes the file. Missing file is a no-op.
func Consume() error {
	data, err := os.ReadFile(AdminFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var admin AdminPayload
	if err := json.Unmarshal(data, &admin); err != nil {
		return err
	}

	user := &model.User{
		Username: admin.Username,
		Email:    admin.Email,
		Status:   1,
	}
	if err := user.SetPassword(admin.Password); err != nil {
		return err
	}
	if err := dao.NewUserDAO(helper.GetHelper()).Create(user); err != nil {
		return err
	}
	helper.GetHelper().GetLogger().Info("[install_bootstrap] 管理员账户已创建: " + admin.Username)
	return os.Remove(AdminFile)
}
