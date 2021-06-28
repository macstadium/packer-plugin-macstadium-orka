# Packer Plugin for Orka

This is a [Packer Plugin](https://www.packer.io/docs/plugins) to automate building images for [MacStadium Orka](https://www.macstadium.com/orka), a Kubernetes/Docker-based macOS virtualization PaaS/SaaS service by [MacStadium](https://www.macstadium.com/).

## Installation

### Using pre-built releases

#### Using the `packer init` command

Starting from version 1.7, Packer supports a new `packer init` command allowing
automatic installation of Packer plugins. Read the
[Packer documentation](https://www.packer.io/docs/commands/init) for more information.

To install this plugin, copy and paste this code into your Packer configuration .
Then, run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    macstadium-orka = {
      version = ">= 2.3.0"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}
```


#### Manual installation

You can find pre-built binary releases of the plugin [here](https://github.com/macstadium/packer-plugin-macstadium-orka/releases).
Once you have downloaded the latest archive corresponding to your target OS,
uncompress it to retrieve the plugin binary file corresponding to your platform.
To install the plugin, please follow the Packer documentation on
[installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).


### From Sources

If you prefer to build the plugin from sources, clone the GitHub repository
locally and run the command `go build` from the root
directory. Upon successful compilation, a `packer-plugin-macstadium-orka` plugin
binary file can be found in the root directory.
To install the compiled plugin, please follow the official Packer documentation
on [installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).

### Configuration

For more information on how to configure the plugin, please read the
documentation located in the [`docs/`](docs) directory.

## Contributing

* If you think you've found a bug in the code or you have a question regarding
  the usage of this software, please reach out to us by opening an issue in
  this GitHub repository.
* Contributions to this project are welcome: if you want to add a feature or a
  fix a bug, please do so by opening a Pull Request in this GitHub repository.
  In case of feature contribution, we kindly ask you to [open an issue](https://github.com/macstadium/packer-plugin-macstadium-orka/issues) to
  discuss it beforehand.

* For more information on contributing, please view [`CONTRIBUTING.md`](CONTRIBUTING.md) 

## Original Author / License

This plugin is "very-loosely" based-on and took inspiration from the [Packer Null Builder] and [Packer LXD Plugin].  

* Written by [Farley Farley] ( farley _at_ neonsurge **dawt** com )
* License Terms: [GNU GPL v3]

[//]: <> (Ignore, below here are links for ease-of-use above)
[Farley Farley]: https://github.com/andrewfarley
[GNU GPL v3]: https://choosealicense.com/licenses/gpl-3.0/
[Packer Null Builder]: https://github.com/hashicorp/packer/tree/master/builder/null
[Packer LXD Plugin]: https://github.com/hashicorp/packer-plugin-lxd
