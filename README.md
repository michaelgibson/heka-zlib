heka-zlib
=========

Heka Zlib Decoder and Filter.

Zlib decoder and filter for [Mozilla Heka](http://hekad.readthedocs.org/)

ZlibDecoder
===========

The Zlib Decoder takes the payload from a Heka message and decompresses it before passing it on.
If specified, message_fields will be interpolated into Heka Fields.

Config:

- message_fields:
    Subsection defining message fields to populate and the interpolated values
    that should be used. Valid interpolated values are any captured in a JSONPath
    in the message_matcher, and any other field that exists in the message. In
    the event that a captured name overlaps with a message field, the captured
    name's value will be used. Optional representation metadata can be added at
    the end of the field name using a pipe delimiter i.e. ResponseSize|B  =
    "%ResponseSize%" will create Fields[ResponseSize] representing the number of
    bytes.  Adding a representation string to a standard message header name
    will cause it to be added as a user defined field i.e., Payload|json will
    create Fields[Payload] with a json representation
    (see :ref:`field_variables`).

    Interpolated values should be surrounded with `%` signs::

Example:

	[zlib_decoder]
	type = "ZlibDecoder"

	[zlib_decoder.message_fields]
	Type = "%Type%Decoded"
	Zlib = "ok"

It probably does not make sense to use the decoder on it's own since compressing a single message at a time would be counterproductive.
Instead it is likely you would be decoding a payload containing multiple messages.
In that case you would need to use it in combination with something like a "split" decoder inside of the Multidecoder.

	[multi_decoder]
	type = "MultiDecoder"
	order = ['zlib_decoder', 'split_decoder', 'json_decoder']

	[multi_decoder.subs.zlib_decoder]
	type = "ZlibDecoder"

	[multi_decoder.subs.split_decoder]
	type = "SandboxDecoder"
	script_type = "lua"
	filename = "/usr/share/heka/lua_decoders/split_decoder.lua"
	preserve_data = true

	[multi_decoder.subs.json_decoder]
	type = "SandboxDecoder"
	script_type = "lua"
	filename = "/usr/share/heka/lua_decoders/json_decoder.lua"
	preserve_data = true


ZlibFilter
==========

The Zlib Filter aggregates the payloads of multiple Heka messages before compressing them into a single Payload and passing it on.

Config: 

- zlib_tag:
	Since the output of the Payload will be binary after this Filter, you will need some way of identifying the message further down the pipeline.
	This setting creates a new Heka Field called "ZlibTag" and is given the value of this option. Defaults to "compressed"

- flush_interval: 
	Interval at which accumulated payloads should be compressed in milliseconds.
	Defaults to 1000 (i.e. one second)

- flush_bytes:
	Number of payloads that, if processed, will trigger them to be compressed.
	Defaults to 10.

- encoder:
	Since the output of the Payload will be binary after this Filter, you will not get the opportunity to encode the message later.
	This option will run each Payload through the specified encoder prior to compressing.

Example:

	[filter_zlib]
	type = "ZlibFilter"
	message_matcher = "Fields[decoded] == 'True'"
	zlib_tag = "compressed"
	flush_interval = 5000
	flush_bytes = 1000000
	encoder = "encoder_json"


To Build
========

See [Building *hekad* with External Plugins](http://hekad.readthedocs.org/en/latest/installing.html#build-include-externals)
for compiling in plugins.

Edit cmake/plugin_loader.cmake file and add

    add_external_plugin(git https://github.com/michaelgibson/heka-zlib master)

Build Heka:
	. ./build.sh
