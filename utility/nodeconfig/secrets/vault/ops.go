package vault

import (
	"errors"
	"fmt"
)

const pathPrefix = "ww/data/"

func (vault *Vault) Read(subpath string) (map[string]interface{}, error) {

	secretpath := pathPrefix + subpath
	LOGGER.Infof("Reading secret from %v", secretpath)
	//read
	secret, err := vault.client.Logical().Read(secretpath)
	if err != nil {
		LOGGER.Errorf("Read error: %v", err)
		return nil, err
	}
	if secret == nil {
		LOGGER.Error("Secret not found")
		return nil, errors.New("Secret not found")
	}

	result := secret.Data["data"].(map[string]interface{})

	return result, nil

	/*


		meta, _ := json.Marshal(secret.Data["metadata"])
		LOGGER.Info("metedata", string(meta))
	*/
}

func (vault *Vault) Create(input map[string]interface{}, subpath string) error {

	secretpath := pathPrefix + subpath
	LOGGER.Infof("Creating Secrets at %v", secretpath)

	newSecret := make(map[string]interface{})
	newSecret["data"] = input

	_, err := vault.client.Logical().Write(secretpath, newSecret)
	if err != nil {
		LOGGER.Errorf("Create error: %v", err)
		return err
	}

	LOGGER.Info("Create Success")
	return nil
}

func (vault *Vault) Append(input map[string]interface{}, subpath string) error {

	secretpath := pathPrefix + subpath

	LOGGER.Infof("Appending Secrets to %v", secretpath)
	secret, err := vault.Read(subpath)
	if err != nil {
		LOGGER.Errorf("%v", err)
		return nil
	}

	for key, value := range input {
		if _, found := secret[key]; found {
			msg := fmt.Sprintf("Key %v already exists in secret %v", key, secretpath)
			return errors.New(msg)
		} else {
			secret[key] = value
		}
	}

	newSecret := make(map[string]interface{})
	newSecret["data"] = secret

	/*
		option := make(map[string]interface{})
		option["cas"] = 1
		newSecret["options"] = option
	*/
	_, err = vault.client.Logical().Write(secretpath, newSecret)
	if err != nil {
		LOGGER.Errorf("Append error: %v", err)
		return err
	}

	LOGGER.Info("Append Success")
	return nil
}

func (vault *Vault) Update(input map[string]interface{}, subpath string) error {
	secretpath := pathPrefix + subpath

	LOGGER.Infof("Updating Secrets at %v", secretpath)

	secret, err := vault.Read(subpath)
	if err != nil {
		LOGGER.Errorf("%v", err)
		return nil
	}

	for key, value := range input {
		if _, found := secret[key]; found {
			secret[key] = value
		} else {
			msg := fmt.Sprintf("Key %v not found in secret %v", key, secretpath)
			return errors.New(msg)
		}
	}

	newSecret := make(map[string]interface{})
	newSecret["data"] = secret

	_, err = vault.client.Logical().Write(secretpath, newSecret)
	if err != nil {
		LOGGER.Errorf("Update error: %v", err)
		return err
	}

	LOGGER.Info("Update Success")
	return nil
}

func (vault *Vault) Override(input map[string]interface{}, subpath string) error {

	secretpath := pathPrefix + subpath
	LOGGER.Infof("Overriding Secrets at %v", secretpath)

	newSecret := make(map[string]interface{})
	newSecret["data"] = input

	_, err := vault.client.Logical().Write(secretpath, newSecret)
	if err != nil {
		LOGGER.Errorf("Override error: %v", err)
		return err
	}

	LOGGER.Info("Override Success")
	return nil
}

func (vault *Vault) Delete(subpath string) error {

	secretpath := pathPrefix + subpath
	LOGGER.Infof("Deleting Secrets at %v", secretpath)

	_, err := vault.client.Logical().Delete(secretpath)
	if err != nil {
		LOGGER.Errorf("Delete error: %v", err)
		return err
	}

	LOGGER.Info("Delete Success")
	return nil
}
