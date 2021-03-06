heka-zlib
=========

Zlib decoder and filter plugins for [Mozilla Heka](http://hekad.readthedocs.org/)

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

See: https://github.com/michaelgibson/heka-stream-aggregator/blob/master/stream_splitter_decoder.go

	[multi_decoder]
	type = "MultiDecoder"
	order = ['zlib_decoder', 'split_decoder', 'json_decoder']

	[multi_decoder.subs.zlib_decoder]
	type = "ZlibDecoder"

	[multi_decoder.subs.split_decoder]
	type = "SplitDecoder"
	[split_decoder.message_fields]
	Split = "ok"

	[multi_decoder.subs.json_decoder]
	type = "SandboxDecoder"
	script_type = "lua"
	filename = "/usr/share/heka/lua_decoders/json_decoder.lua"
	preserve_data = true


ZlibEncoder
==========
Encodes the payload of a pack into a zlib stream that may be decoded using ZlibDecoder

Config:

- append_newlines (bool, optional):
	Specifies whether or not a newline character (i.e. `\n`) will be appended
	to the captured message payload before serialization. Defaults to true.

- prefix_ts (bool, optional):
	Specifies whether a timestamp will be prepended to the captured message
	payload before serialization. Defaults to false.

- ts_from_message (bool, optional):
	If true, the prepended timestamp will be extracted from the message that
	is being processed. If false, the prepended timestamp will be generated by
	the system clock at the time of message processing. Defaults to true. This
	setting has no impact if `prefix_ts` is set to false.

- ts_format (string, optional):
	Specifies the format that should be used for prepended timestamps, using
	Go's standard `time format specification strings
	<http://golang.org/pkg/time/#pkg-constants>`_. Defaults to
	`[2006/Jan/02:15:04:05 -0700]`. If the specified format string does not
	end with a space character, then a space will be inserted between the
	formatted timestamp and the payload.

Example

	[zlib_encoder]
	type = "ZlibEncoder"
	append_newlines = false
	prefix_ts = true
	ts_format = "2006/01/02 3:04:05PM MST"


To Build
========

See [Building *hekad* with External Plugins](http://hekad.readthedocs.org/en/latest/installing.html#build-include-externals)
for compiling in plugins.

Edit cmake/plugin_loader.cmake file and add

    add_external_plugin(git https://github.com/michaelgibson/heka-zlib master)

Build Heka:
	. ./build.sh
