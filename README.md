# Packer Builder for Orka

This is a [Packer Builder] to automate building images for [MacStadium Orka] a Kubernetes/Docker-based macOS virtualization SaaS service by [MacStadium].

![example screenshot](./images/screenshot1.jpg)

### Compatibility

For this plugin to function you need to have at least Packer 1.6.0 installed and Orka CLI 1.3.0.

 * [Packer Downloads] - 1.6.0+
 * [Orka CLI Downloads] - 1.3.0+

## Install / Setup

1. Install [Packer](https://www.packer.io/downloads.html)
2. Install [Orka CLI](https://orkadocs.macstadium.com/docs/downloads)
3. Setup Orka CLI - See: [Orka Setup Guide]
3. Download the [Latest Release] of this plugin.
4. Unzip the plugin binaries to a location where [Packer] will detect them at run-time, such as any of the following:
    * The directory where the [packer] binary is.
    * The `~/.packer.d/plugins` directory.
    * The current working directory.
5. Change to a directory where you have [packer] templates, and run as usual.

## Packer Builder Configuration

```json
{
  "builders": [{
    "type": "macstadium-orka",
    "source_image": "name-of-image-from-vm-images-list",
    "image_name": "destination-image-name"
  }]
}
```
---

* `type` _(string)_ **(required)**

Must be `macstadium-orka`

* `source_image` _(string)_ **(required)**

This is the source image (vm config) we will be using to launch the VM from.  This should be an entry from `orka vm configs`.  If you don't have one or need to create one, here's an example.  The "base-image" in the below command is looked up from `orka image list`.  See [examples below](#Example-Commands)

* `image_name` _(string)_ (optional)

This is the destination name of the image that will be created.  The image will be located inside `orka image list` when completed.  If not specified this will be autogenerated to the following: `packer-{{unix timestamp}}`

### Development / Internal Options

If you're NOT a dev working on this software you can ignore the following.

But, if you are building/editing/updating this software, you may want to turn on most or all of the following options.  See the [examples/macos-catalina.json](./examples/macos-catalina.json).

* `simulate_create` _(boolean)_ (optional) _*- for devs*_

This is for internal development purposes, to prevent having to constantly create VMs for testing/development of this plugin.  Not for normal use.  The simulated "example" is hardcoded into the code on line 90 of [builder/orka/step_orka_create.go](./builder/orka/step_orka_create.go).  Feel free to edit during local development for your environment, but do not submit a commit/MR editing this please.

* `do_not_delete` _(boolean)_ (optional) _*- for devs*_

By default this plugin automatically deletes the VM afterwards if all scripts ran successfully.  AFAIK there is no mechanism built-into Packer to "force" it to not delete the source VM afterwards, so this fills that gap.  This is useful for debugging builds that are being weird, but is generally not intended for noraml use.  You should just use the resulting image.

* `do_not_image` _(boolean)_ (optional) _*- for devs*_

By default this plugin automatically creates an image of the VM after any provisioning steps.  AFAIK there is no mechanism built-into Packer to "force" it to not image the VM afterwards, so this fills that gap.  This is useful for debugging builds that are being weird, but is generally not intended for noraml use.


## Information Notes / Gotchas

[MacStadium Orka] base images have SSH enabled by default and the username/password is `admin:admin` because they are within' a private network by default.  So this plugin has those credentials hardcoded by default, but you can of course customize the communicator.  See the options from the [SSH Communicator].

## Example Orka Commands

These aren't directly related to this plugin, exactly, but they're a bit of a simplified guide to get you started.  For a more full guide see: [Orka Setup Guide].

```bash
# Create a config which is used for source_image above
orka vm create-config -v macos-catalina-10-15-5 -c 3 --C 3 --vnc --base-image macos-catalina-10.15.5.img -y

# In essence, this plugin automates running the following 3 commands...
# Start a VM (using the config above)
orka vm deploy -v macos-catalina-10-15-5 --vnc -y

# Save a VM's disk to a disk image
orka image save -v <vmid-here-from-orka-vm-list> -b <destination-image-name> -y

# Stop and remove a VM
orka vm delete --vm base-image-catalina -y

# Once you're done working with an image and want to delete
orka image delete --image <destination-image-name> -y

# Or alternatively, use that image in a future config (orka vm create-config) to launch future VMs based on that image
```

## Changelog / History

[1.0.0] - Initial Release, basic functionality wrapper around Orka CLI

## Development / Bugs / Support

If you want to help contribute to this plugin or find a bug, please [file an issue] on Github, and/or submit me a PR.  This is Open Source, so don't expect me to fix bugs immediately, but I'll try my best to reasonably support this plugin.  Contributors are always welcome though.

To get a development environment up will need a recent golang installed and setup.  Then with a single make command below command it will build, install, and try to run the example at [examples/macos-catalina.json](./examples/macos-catalina.json).  You may need to edit this file to have the source VM config that you have locally, as it is hardcoded to my environment's image `macos-catalina-10-15-5` at the moment.

```bash
make fresh
```

Finally, if you want to support me or this project in some way, please donate to a local animal shelter or makerspace or one of many open-source projects that you rely on regularly.

## Todo (in no particular order)

These are a list of things that are pending to accomplish within' this repo.  Contributors welcome, I might do some of these also, eventually.

 * Add tests (should do this first though, as this is important to keep this plugin functional and debugging issues)
 * Add image management features into the builder instead of having to manage out-of-band with orka cli (eg: delete image)
 * Migrate to the (as of yet) undocumented Orka API.  Maybe MacStadium will help and provide me some documentation?  This will simplify the code greatly, and remove the reliance on the Orka CLI to be pre-installed and pre-configured.  Although, arguably this may make this more complicated because then you need to specify the API Key/Token/Account information into this plugin.
 * Clean up / improve code / catch more sharp edges and edge-cases, deal with any issues filed on Github.

## Original Author / License

**Please Note:** I am not associated with, affiliated, or tied to MacStadium in any way.  They did not endorse the creation or support of this plugin.  This software was written as a personal open-source contribution to the community.  If you create work based on this work, please attribute that your work was based on mine.

This plugin is "very-loosely" based on and took inspiration from the [Packer Null Builder], [Packer LXD Builder], and the [Packer Builder Veertu Anka].

* Written by [Farley Farley] ( farley _at_ neonsurge **dawt** com )
* License Terms: [GNU GPL v3]




[//]: <> (Ignore, below here are links for ease-of-use above)
[Packer]: https://www.packer.io/
[Packer Builder]: https://www.packer.io/docs/extending/custom-builders.html
[MacStadium Orka]: https://www.macstadium.com/orka
[Orka]: https://www.macstadium.com/orka
[MacStadium]: https://www.macstadium.com
[Packer Downloads]: https://www.packer.io/downloads.html
[Orka CLI Downloads]: https://orkadocs.macstadium.com/docs/downloads
[Orka Setup Guide]: https://orkadocs.macstadium.com/docs/quick-start
[Latest Release]: https://github.com/andrewfarley/packer-builder-macstadium-orka/releases
[Farley Farley]: https://github.com/andrewfarley
[GNU GPL v3]: https://choosealicense.com/licenses/gpl-3.0/
[1.0.0]: https://github.com/andrewfarley/packer-builder-macstadium-orka/releases/tag/v1.0.0
[SSH Communicator]: https://www.packer.io/docs/communicators/ssh
[Packer Builder Veertu Anka]: https://github.com/veertuinc/packer-builder-veertu-anka
[Packer Null Builder]: https://github.com/hashicorp/packer/tree/master/builder/null
[Packer LXD Builder]: https://github.com/hashicorp/packer/tree/master/builder/lxd
[file an issue]: https://github.com/AndrewFarley/packer-builder-macstadium-orka/issues