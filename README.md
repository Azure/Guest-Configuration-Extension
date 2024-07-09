[![Build Status](https://travis-ci.org/Azure/Guest-Configuration-Extension.svg?branch=master)](https://travis-ci.org/Azure/Guest-Configuration-Extension)

# Guest Configuration Extension for Linux

The Guest Configuration Extension for Linux configures the Guest Configuration 
Agent on VMs. Together, they allow a customer to run In-Guest Policy on their 
VMs, which gives the customer the ability to monitor their system and security 
policies on their machines. In-Guest Policy for Linux currently uses policies 
found on Chef InSpec.

## 1. Deployment to a Virtual Machine

To deploy the Guest Configuration Extension for Linux onto your machine, run:

    $ az vm extension set --resource-group <resource-group> --vm-name <vm-name> \
        --name ConfigurationForLinux --publisher Microsoft.GuestConfiguration

## 2. Commands Guide

The Guest Configuration Extension for Linux supports five commands -- install, enable,
update, disable, and uninstall. To run any of these commands, go to the path: `/var/lib/waagent/Microsoft.GuestConfiguration.ConfigurationForLinux-<version>/bin`, 
and run:

    $ guest-configuration-shim <command name>

##### Install
`Install` does not do anything in itself, but when the Guest Configuration Extension is 
installed, `Enable` will be called immediately aftwards. 

##### Enable
`Enable` handles the configuration of the Guest Configuration Agent. It handles the unzipping of the Agent 
package and then installs and enables the Agent. 

##### Update
`Update` will update the Agent Service to the new Extension. It parses the path of the old Agent, and gives it to the new Agent, so that the agent
can update the service endpoint.

##### Disable
`Disable` disables the agent and returns the status to the user. 

##### Uninstall
`Uninstall` uninstalls the agent, and then the Guest Agent removes everything from the box. 


## 3. Troubleshooting

The agent is downloaded to a path like: `/var/lib/waagent/Microsoft.GuestConfiguration.ConfigurationForLinux-<version>/GCAgent/GC`
and the Agent output is saved to `stdout` and `stderr` files in this directory. Please read
these files to find out output from the agent.

You can find the logs for the extension at a path like: `/var/log/azure/Microsoft.GuestConfiguration.ConfigurationForLinux`.

Please open an issue on this GitHub repository if you encounter problems that
you could not debug with these log files.

## 4. Future Plans

The Guest Configuration Extension for Linux will be made cross-platform to
support both Linux and Windows VMs. It will support all Azure endorsed distributions.

-----
This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
