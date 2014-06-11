/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2012
# the Initial Developer. All Rights Reserved.
#
# ***** END LICENSE BLOCK *****/

package zlib

import (
    "compress/zlib"
    . "github.com/mozilla-services/heka/pipeline"
    "bytes"
)

type ZlibDecoderConfig struct {
    // Keyed to the message field that should be filled in, the value will be
    // interpolated so it can use capture parts from the message match.
    MessageFields MessageTemplate `toml:"message_fields"`

}

type ZlibDecoder struct {
	dRunner         DecoderRunner
    MessageFields   MessageTemplate
}

func (ld *ZlibDecoder) ConfigStruct() interface{} {

	return &ZlibDecoderConfig{
	}
}

func (ld *ZlibDecoder) Init(config interface{}) (err error) {
    conf := config.(*ZlibDecoderConfig)
    ld.MessageFields = make(MessageTemplate)

    if conf.MessageFields != nil {
            for field, action := range conf.MessageFields {
                    ld.MessageFields[field] = action
            }
    }


	return
}

// Heka will call this to give us access to the runner.
func (ld *ZlibDecoder) SetDecoderRunner(dr DecoderRunner) {
	ld.dRunner = dr
}


// Runs the message payload against decoder's map of JSONPaths. If
// there's a match, the message will be populated based on the
// decoder's message template, with capture values interpolated into
// the message template values.
func (ld *ZlibDecoder) Decode(pack *PipelinePack) (packs []*PipelinePack, err error) {
    b := bytes.NewBufferString(pack.Message.GetPayload())
    r, err := zlib.NewReader(b)
    buf := new(bytes.Buffer)

    if b.Len() > 0 {
        buf.ReadFrom(r)
        s := buf.String()
        pack.Message.SetPayload(s)
    }
    
    packs = []*PipelinePack{pack}

    return
}

func init() {
        RegisterPlugin("ZlibDecoder", func() interface{} {
                return new(ZlibDecoder)
        })
}

