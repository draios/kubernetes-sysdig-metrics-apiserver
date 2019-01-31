The playground is designed to run `k8s-sysdig-adapter` locally using virtual
machines. We provision a Kubernetes cluster with a single controller node and a
configurable number of worker nodes.

This playground borrows some ideas from other Vagrant-based environments I've
found in projects like [kubevirt][1], [k8s-snowflake][2] or
[kubernetes-ansible-vagrant][3]. Thank you!

### Requirements

**Please use the most recent versions of Vagrant and VirtualBox!**

Each node is assigned 2GB of RAM, including the controller node.

### Installation

To create the environment run:

    ./setup.sh

Optionally, you can define the number of worker nodes and/or enable the debug
mode so the shell scripts become verbose, e.g.:

    env DEBUG=1 WORKERS=2 ./setup.sh

The default is to deplay a single worker which is enough in most cases.
[Issue #1][7] needs to be considered if you're deploying more than one worker.

This script is going to do a few things for us:

- It runs `vagrant up` for us to provision the virtual machines, which uses
[bootstrap.sh][4] to install the necessary packages and apply a few tweaks on
each node so the Kubernetes components can run properly.

- It runs [kubeadm.sh][5] which uses [kubeadm][6] to set up the Kubernetes
cluster. The kubeconfig generated is copied in a temporary directory of the host
machine and its location is printed before the script ends so you can use it to
operate the cluster.

If the script completes successfully you should see something like the
following:

```
We're done, thank you for waiting!
kubectl config available in /tmp/tmp.8nJG5UCHA7/config

Usage example:
$   export KUBECONFIG=/tmp/tmp.8nJG5UCHA7/config
$   kubectl get nodes
-------------------------------------------------------------------------
```

Copy the `config` file somewhere else if your temporary folder is not
persistent. You can optionally consolidate the output with your local
`.kube/config` - we thought it would be safer if you do that manually!

### Check the status of the cluster

Let's make sure that the cluster is running as expected. We're going to export
a custom `KUBECONFIG` environment string so we don't have to pass the location
on every command.

    $ export KUBECONFIG=/tmp/tmp.8nJG5UCHA7/config

Let's confirm that all the nodes are listed as ready:

    $ kubectl get nodes
    NAME              STATUS    ROLES     AGE       VERSION
    controller-node   Ready     master    4m        v1.10.0
    worker-node-1     Ready     <none>    4m        v1.10.0
    worker-node-2     Ready     <none>    4m        v1.10.0

Now let's confirm that the core components are in a healthy state:

    $ kubectl get componentstatuses
    NAME                 STATUS    MESSAGE              ERROR
    controller-manager   Healthy   ok
    scheduler            Healthy   ok
    etcd-0               Healthy   {"health": "true"}

You're ready! :tada:

[1]: https://github.com/kubevirt/kubevirt
[2]: https://github.com/jessfraz/k8s-snowflake
[3]: https://github.com/errordeveloper/kubernetes-ansible-vagrant
[4]: ./mixins/bootstrap.sh
[5]: ./misins/kubeadm.sh
[6]: https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm/
[7]: https://github.com/dcberg/kubernetes-sysdig-metrics-apiserver/issues/1
