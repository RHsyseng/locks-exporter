# locks-exporter
Prometheus exporter for system file lock counts

# Usage
The `locks-exporter` can usually run without any additional configuration.

```
usage: locks-exporter [<flags>]

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
      --lock.procfsPath="/proc"  Path to procfs filesystem.
      --log.level="info"         Log level.
      --web.listen-address=":9102"  
                                 Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"  
                                 Path under which to expose metrics.
      --version                  Show application version.
```

The exp a `locks_container_file_locks` gauge metric for the number of files locked by each cri-o container on the host. An example output from the `/metrics` endpoint is below:
```
# HELP locks_container_file_locks Number of file locks held by processes in container
# TYPE locks_container_file_locks gauge
locks_container_file_locks{container="etcd",namespace="openshift-etcd",pod="etcd-master-0"} 1
locks_container_file_locks{container="nbdb",namespace="openshift-ovn-kubernetes",pod="ovnkube-master-fhjkg"} 2
locks_container_file_locks{container="northd",namespace="openshift-ovn-kubernetes",pod="ovnkube-master-fhjkg"} 1
locks_container_file_locks{container="ovn-controller",namespace="openshift-ovn-kubernetes",pod="ovnkube-node-z8vng"} 1
locks_container_file_locks{container="ovnkube-master",namespace="openshift-ovn-kubernetes",pod="ovnkube-master-fhjkg"} 1
locks_container_file_locks{container="sbdb",namespace="openshift-ovn-kubernetes",pod="ovnkube-master-fhjkg"} 2
```

# Deployment
The exporter can be deployed in an OpenShift cluster as a daemonset to collect file lock counts on each node. The exporter includes a `PodMonitor` to enable scraping by a Prometheus instance.

1. Create a project for the exporter
```
$ oc new-project locks
```

2. Deploy the resources for the locks-exporter. This includes the necessary RBAC configuration to allow the exporter to use the `hostaccess` Security Context Constraint, which is needed to read the procfs of the node.
```
$ oc create -f k8s/exporter.yml
serviceaccount/locks-exporter created
role.rbac.authorization.k8s.io/locks-exporter-hostaccess created
rolebinding.rbac.authorization.k8s.io/locks-exporter-hostaccess created
daemonset.apps/locks-exporter created
podmonitor.monitoring.coreos.com/locks-exporter created
```

3. If using the OpenShift monitoring stack, check to ensure [monitoring for user-defined projects](https://docs.openshift.com/container-platform/4.9/monitoring/enabling-monitoring-for-user-defined-projects.html) is enabled.
```
$ oc describe cm cluster-monitoring-config -n openshift-monitoring
Name:         cluster-monitoring-config
Namespace:    openshift-monitoring
Labels:       <none>
Annotations:  <none>

Data
====
config.yaml:
----
enableUserWorkload: true
```

4. You should now be able to query Prometheus for the `locks_container_file_locks` metric.
