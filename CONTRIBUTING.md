## Development / Bugs / Support

If you want to help contribute to this plugin or find a bug, please [file an issue] on Github, and/or submit a PR.

To start developing the plugin:
- Make sure [golang](https://golang.org/) is installed and setup
- Edit [examples/orka.pkr.hcl](./examples/orka.pkr.hcl) to match the configuration you want to test with
- Run `make fresh` or `make rebuild` to build, install, and try to run the example at [examples/orka.pkr.hcl](./examples/orka.pkr.hcl)

## Todo (in no particular order)

These are a list of things that are pending to accomplish within' this repo.  Contributors welcome, I might do some of these also, eventually.

 * Add more tests 
 * Add image management features into the builder instead of having to manage out-of-band with orka cli (eg: delete image)
 * Consider implementing for this plugin to automatically create a VM Config (before launching a VM, and possibly after tied to the image just created)
 * Add the ability for this plugin to be able to scan configs and/or images available and automatically use the first one it finds, possibly creating a new config specifically for the image?  
 * Improve the JSON parsing code.  See: line 137-176 of [builder/orka/step_orka_create.go](./builder/orka/step_orka_create.go) and look at the function `ExtractIPHost` in that file as well.  Could be much improved and hopefully simplified.  Contributors welcome!!!
 * Clean up / improve code / catch more sharp edges and edge-cases, deal with any issues filed on Github.
 * One day, get this merged upstream into Packer as an official packer plugin.  Additional info [here](https://www.packer.io/docs/plugins/packer-integration-program)



[//]: <> (Ignore, below here are links for ease-of-use above)
[file an issue]: https://github.com/macstadium/packer-plugin-macstadium-orka/issues