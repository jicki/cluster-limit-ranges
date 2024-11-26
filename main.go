package main

import (
	"flag"
	"github.com/jicki/cluster-limit-ranges/api/v1"
	"github.com/jicki/cluster-limit-ranges/controllers"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server" // 引入 metrics server 的新包
)

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.Parse()

	// 设置日志记录器
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Metrics: server.Options{
			BindAddress: metricsAddr, // 配置 metrics server 的绑定地址
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "jicki.cn.clusterlimit",
	})
	if err != nil {
		os.Exit(1)
	}

	// 添加 CRD 资源到 scheme
	if err := v1.AddToScheme(mgr.GetScheme()); err != nil {
		os.Exit(1)
	}

	// 设置 Reconciler
	if err := (&controllers.ClusterLimitReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		os.Exit(1)
	}

	// 健康检查和就绪检查
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		os.Exit(1)
	}

	// 启动 Manager
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
