package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newAttachDBCmd() *cobra.Command {
	var (
		name, namespace, user, password, host, db string
		port int
	)
	cmd := &cobra.Command{
		Use:   "attach-db",
		Short: "Attach Postgres by creating a Secret with DATABASE_URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" || host == "" || db == "" { return fmt.Errorf("name, host, db required") }
			ns := namespace; if ns == "" { ns = "default" }
			client, ns2, err := getClient(ns); if err != nil { return err }
			url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, db)
			sec := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name + "-db", Namespace: ns2}, StringData: map[string]string{"DATABASE_URL": url}}
			_, err = client.CoreV1().Secrets(ns2).Get(context.Background(), sec.Name, metav1.GetOptions{})
			if err == nil { _, err = client.CoreV1().Secrets(ns2).Update(context.Background(), sec, metav1.UpdateOptions{}); return err }
			_, err = client.CoreV1().Secrets(ns2).Create(context.Background(), sec, metav1.CreateOptions{})
			return err
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "App name")
	cmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace")
	cmd.Flags().StringVar(&user, "user", "app", "DB user")
	cmd.Flags().StringVar(&password, "password", "changeme", "DB password")
	cmd.Flags().StringVar(&host, "host", "", "DB host")
	cmd.Flags().IntVar(&port, "port", 5432, "DB port")
	cmd.Flags().StringVar(&db, "db", "", "DB name")
	return cmd
}

func newAttachRedisCmd() *cobra.Command {
	var (
		name, namespace, host, password string
		port int
	)
	cmd := &cobra.Command{
		Use:   "attach-redis",
		Short: "Attach Redis by creating a Secret with REDIS_URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" || host == "" { return fmt.Errorf("name and host required") }
			ns := namespace; if ns == "" { ns = "default" }
			client, ns2, err := getClient(ns); if err != nil { return err }
			url := fmt.Sprintf("redis://:%s@%s:%d", password, host, port)
			sec := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name + "-redis", Namespace: ns2}, StringData: map[string]string{"REDIS_URL": url}}
			_, err = client.CoreV1().Secrets(ns2).Get(context.Background(), sec.Name, metav1.GetOptions{})
			if err == nil { _, err = client.CoreV1().Secrets(ns2).Update(context.Background(), sec, metav1.UpdateOptions{}); return err }
			_, err = client.CoreV1().Secrets(ns2).Create(context.Background(), sec, metav1.CreateOptions{})
			return err
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "App name")
	cmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace")
	cmd.Flags().StringVar(&host, "host", "", "Redis host")
	cmd.Flags().IntVar(&port, "port", 6379, "Redis port")
	cmd.Flags().StringVar(&password, "password", "", "Redis password")
	return cmd
}

func newSecretsCmd() *cobra.Command {
	var (
		name, namespace string
		pairs []string
	)
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Create/update Secret from key=value pairs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" { return fmt.Errorf("name required") }
			ns := namespace; if ns == "" { ns = "default" }
			client, ns2, err := getClient(ns); if err != nil { return err }
			data := map[string]string{}
			for _, p := range pairs { kv := strings.SplitN(p, "=", 2); if len(kv) != 2 { return fmt.Errorf("invalid pair: %s", p) }; data[kv[0]] = kv[1] }
			sec := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns2}, StringData: data}
			_, err = client.CoreV1().Secrets(ns2).Get(context.Background(), name, metav1.GetOptions{})
			if err == nil { _, err = client.CoreV1().Secrets(ns2).Update(context.Background(), sec, metav1.UpdateOptions{}); return err }
			_, err = client.CoreV1().Secrets(ns2).Create(context.Background(), sec, metav1.CreateOptions{})
			return err
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Secret name")
	cmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace")
	cmd.Flags().StringSliceVar(&pairs, "from-literal", []string{}, "key=value")
	return cmd
}
