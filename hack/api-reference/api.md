<p>Packages:</p>
<ul>
<li>
<a href="#hcloud.provider.extensions.gardener.cloud%2fv1alpha1">hcloud.provider.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="hcloud.provider.extensions.gardener.cloud/v1alpha1">hcloud.provider.extensions.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the HCloud provider API resources.</p>
</p>
Resource Types:
<ul><li>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudProfileConfig">CloudProfileConfig</a>
</li><li>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.ControlPlaneConfig">ControlPlaneConfig</a>
</li></ul>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudProfileConfig">CloudProfileConfig
</h3>
<p>
<p>CloudProfileConfig contains provider-specific configuration that is embedded into Gardener&rsquo;s <code>CloudProfile</code>
resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
hcloud.provider.extensions.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>CloudProfileConfig</code></td>
</tr>
<tr>
<td>
<code>regions</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.RegionSpec">
[]RegionSpec
</a>
</em>
</td>
<td>
<p>Regions is the specification of regions and zones topology</p>
</td>
</tr>
<tr>
<td>
<code>machineImages</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineImages">
[]MachineImages
</a>
</em>
</td>
<td>
<p>MachineImages is the list of machine images that are understood by the controller. It maps
logical names and versions to provider-specific identifiers.</p>
</td>
</tr>
<tr>
<td>
<code>defaultClassStoragePolicyName</code></br>
<em>
string
</em>
</td>
<td>
<p>DefaultClassStoragePolicyName is the name of the HCloud storage policy to use for the &lsquo;default-class&rsquo; storage class</p>
</td>
</tr>
<tr>
<td>
<code>machineTypeOptions</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineTypeOptions">
[]MachineTypeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MachineTypeOptions is the list of machine type options to set additional options for individual machine types.</p>
</td>
</tr>
<tr>
<td>
<code>dockerDaemonOptions</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.DockerDaemonOptions">
DockerDaemonOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DockerDaemonOptions contains configuration options for docker daemon service</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.ControlPlaneConfig">ControlPlaneConfig
</h3>
<p>
<p>ControlPlaneConfig contains configuration settings for the control plane.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
hcloud.provider.extensions.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>ControlPlaneConfig</code></td>
</tr>
<tr>
<td>
<code>cloudControllerManager</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudControllerManagerConfig">
CloudControllerManagerConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CloudControllerManager contains configuration settings for the cloud-controller-manager.</p>
</td>
</tr>
<tr>
<td>
<code>loadBalancerClasses</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.CPLoadBalancerClass">
[]CPLoadBalancerClass
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LoadBalancerClasses lists the load balancer classes to be used.</p>
</td>
</tr>
<tr>
<td>
<code>loadBalancerSize</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>LoadBalancerSize can override the default of the NSX-T load balancer size (&ldquo;SMALL&rdquo;, &ldquo;MEDIUM&rdquo;, or &ldquo;LARGE&rdquo;) defined in the cloud profile.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.CPLoadBalancerClass">CPLoadBalancerClass
</h3>
<p>
(<em>Appears on:</em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.ControlPlaneConfig">ControlPlaneConfig</a>)
</p>
<p>
<p>CPLoadBalancerClass provides the name of a load balancer</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>ipPoolName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPPoolName is the name of the NSX-T IP pool.</p>
</td>
</tr>
<tr>
<td>
<code>tcpAppProfileName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TCPAppProfileName is the profile name of the load balaner profile for TCP</p>
</td>
</tr>
<tr>
<td>
<code>udpAppProfileName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>UDPAppProfileName is the profile name of the load balaner profile for UDP</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudControllerManagerConfig">CloudControllerManagerConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.ControlPlaneConfig">ControlPlaneConfig</a>)
</p>
<p>
<p>CloudControllerManagerConfig contains configuration settings for the cloud-controller-manager.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>featureGates</code></br>
<em>
map[string]bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>FeatureGates contains information about enabled feature gates.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.DockerDaemonOptions">DockerDaemonOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudProfileConfig">CloudProfileConfig</a>)
</p>
<p>
<p>DockerDaemonOptions contains configuration options for Docker daemon service</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>httpProxyConf</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTPProxyConf contains HTTP/HTTPS proxy configuration for Docker daemon</p>
</td>
</tr>
<tr>
<td>
<code>insecureRegistries</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>InsecureRegistries adds the given registries to Docker on the worker nodes
(see <a href="https://docs.docker.com/registry/insecure/">https://docs.docker.com/registry/insecure/</a>)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineImageVersion">MachineImageVersion
</h3>
<p>
(<em>Appears on:</em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineImages">MachineImages</a>)
</p>
<p>
<p>MachineImageVersion contains a version and a provider-specific identifier.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
<p>Version is the version of the image.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineImages">MachineImages
</h3>
<p>
(<em>Appears on:</em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudProfileConfig">CloudProfileConfig</a>, 
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.RegionSpec">RegionSpec</a>)
</p>
<p>
<p>MachineImages is a mapping from logical names and versions to provider-specific identifiers.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the logical name of the machine image.</p>
</td>
</tr>
<tr>
<td>
<code>versions</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineImageVersion">
[]MachineImageVersion
</a>
</em>
</td>
<td>
<p>Versions contains versions and a provider-specific identifier.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineTypeOptions">MachineTypeOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudProfileConfig">CloudProfileConfig</a>)
</p>
<p>
<p>MachineTypeOptions defines additional VM options for an machine type given by name</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the machine type</p>
</td>
</tr>
<tr>
<td>
<code>extraConfig</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtraConfig allows to specify additional VM options.
e.g. sched.swap.vmxSwapEnabled=false to disable the VMX process swap file</p>
</td>
</tr>
</tbody>
</table>
<h3 id="hcloud.provider.extensions.gardener.cloud/v1alpha1.RegionSpec">RegionSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.CloudProfileConfig">CloudProfileConfig</a>)
</p>
<p>
<p>RegionSpec specifies the topology of a region and its zones.
A region consists of a Vcenter host, transport zone and optionally a data center.
A zone in a region consists of a data center (if not specified in the region), a computer cluster,
and optionally a resource zone or host system.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the region</p>
</td>
</tr>
<tr>
<td>
<code>machineImages</code></br>
<em>
<a href="#hcloud.provider.extensions.gardener.cloud/v1alpha1.MachineImages">
[]MachineImages
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MachineImages is the list of machine images that are understood by the controller. If provided, it overwrites the global
MachineImages of the CloudProfileConfig</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
