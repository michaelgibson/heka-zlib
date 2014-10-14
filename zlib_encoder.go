/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2014
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Michael Gibson (michael.gibson79@gmail.com)
#
# ***** END LICENSE BLOCK *****/

package encoders

import (
    . "github.com/mozilla-services/heka/pipeline"
	"strings"
	"time"
    "bytes"
    "compress/zlib"
)

type ZlibEncoder struct {
	config *ZlibEncoderConfig
}

type ZlibEncoderConfig struct {
	AppendNewlines bool   `toml:"append_newlines"`
	PrefixTs       bool   `toml:"prefix_ts"`
	TsFromMessage  bool   `toml:"ts_from_message"`
	TsFormat       string `toml:"ts_format"`
}

func (pe *ZlibEncoder) ConfigStruct() interface{} {
	return &ZlibEncoderConfig{
		AppendNewlines: true,
		TsFormat:       "[2006/Jan/02:15:04:05 -0700]",
		TsFromMessage:  true,
	}
}

func (pe *ZlibEncoder) Init(config interface{}) (err error) {
	pe.config = config.(*ZlibEncoderConfig)
	if !strings.HasSuffix(pe.config.TsFormat, " ") {
		pe.config.TsFormat += " "
	}
	return
}

func (pe *ZlibEncoder) Encode(pack *PipelinePack) (output []byte, err error) {
//func (pe *ZlibEncoder) Encode(pack *pipeline.PipelinePack) (output []byte, err error) {
	p := pack.Message.GetPayload()

    var b bytes.Buffer
    w := zlib.NewWriter(&b)
    w.Write([]byte(p))
    w.Close()

    payload := b.String()

	if !pe.config.AppendNewlines && !pe.config.PrefixTs {
		// Just the payload, ma'am.
		output = []byte(payload)
		return
	}

	if !pe.config.PrefixTs {
		// Payload + newline.
		output = make([]byte, 0, len(payload)+1)
		output = append(output, []byte(payload)...)
		output = append(output, '\n')
		return
	}

	// We're using a timestamp.
	var tm time.Time
	if pe.config.TsFromMessage {
		tm = time.Unix(0, pack.Message.GetTimestamp())
	} else {
		tm = time.Now()
	}
	ts := tm.Format(pe.config.TsFormat)

	// Timestamp + payload [+ optional newline].
	l := len(ts) + len(payload)
	output = make([]byte, 0, l+1)
	output = append(output, []byte(ts)...)
	output = append(output, []byte(payload)...)
	if pe.config.AppendNewlines {
		output = append(output, '\n')
	}
	return
}

func init() {
	pipeline.RegisterPlugin("ZlibEncoder", func() interface{} {
		return new(ZlibEncoder)
	})
}
