package usecases

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/custom-database/internal/customdatabase"
	v1 "k8s.io/custom-database/pkg/apis/cusotmdatabase/v1"
)

const (
	SecretVarDbHost     = "DB_HOST"
	SecretVarDbPort     = "DB_PORT"
	SecretVarDbName     = "DB_NAME"
	SecretVarDbUserName = "DB_USERNAME"
	SecretVarDbPassword = "DB_PASSWORD"
)

func newEmptySecret(cd *v1.CustomDatabase) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cd.Spec.SecretName,
			Namespace: cd.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cd, v1.SchemeGroupVersion.WithKind("CustomDatabase")),
			},
			Labels: map[string]string{
				"controller": cd.Name,
			},
		},
	}
}

func secretWithDBInfo(secret *corev1.Secret, dbInfo customdatabase.Entity) *corev1.Secret {
	newSecret := secret.DeepCopy()

	if newSecret.Data == nil {
		newSecret.Data = make(map[string][]byte)
	}
	newSecret.Data[SecretVarDbHost] = []byte(dbInfo.Host.Name)
	newSecret.Data[SecretVarDbPort] = []byte(fmt.Sprintf("%d", dbInfo.Host.Port))
	newSecret.Data[SecretVarDbName] = []byte(dbInfo.Database.Name)
	newSecret.Data[SecretVarDbUserName] = []byte(dbInfo.Database.User)
	newSecret.Data[SecretVarDbPassword] = []byte(dbInfo.Database.Password)

	return newSecret
}
