package zlib

import (
        "fmt"
        "time"
        "github.com/mozilla-services/heka/message"
        . "github.com/mozilla-services/heka/pipeline"
        "bytes"
        "compress/zlib"
        "code.google.com/p/go-uuid/uuid"
        "sync"
)

type ZlibFilter struct {
        *ZlibFilterConfig
        batchChan chan []byte
        backChan  chan []byte
        msgLoopCount uint
}

type ZlibFilterConfig struct {
        FlushInterval uint32 `toml:"flush_interval"`
        FlushBytes    int    `toml:"flush_bytes"`
        ZlibTag       string `toml:"zlib_tag"`
        EncoderName   string `toml:"encoder"`
}

func (f *ZlibFilter) ConfigStruct() interface{} {
        return &ZlibFilterConfig{
        FlushInterval: 1000,
        FlushBytes:    10,
        ZlibTag:       "compressed",
        }
}

func (f *ZlibFilter) Init(config interface{}) (err error) {
        f.ZlibFilterConfig = config.(*ZlibFilterConfig)
        f.batchChan = make(chan []byte)
        f.backChan = make(chan []byte, 2)

        if f.ZlibTag == "" {
            return fmt.Errorf(`A zlib_tag value must be specified for the ZlibTag Field`)
        }

        if f.EncoderName == "" {
            return fmt.Errorf(`An encoder must be specified`)
        }

        return
}

func (f *ZlibFilter) committer(fr FilterRunner, h PluginHelper, wg *sync.WaitGroup) {
        initBatch := make([]byte, 0, 10000)
        f.backChan <- initBatch
        var (
            tag string
            //ok bool
            outBatch []byte
        )
        tag = f.ZlibTag

        for outBatch = range f.batchChan {
                pack := h.PipelinePack(f.msgLoopCount)
                if pack == nil {
                        fr.LogError(fmt.Errorf("exceeded MaxMsgLoops = %d",
                                Globals().MaxMsgLoops))
            break   
                }
                var b bytes.Buffer
                w := zlib.NewWriter(&b)
                w.Write(outBatch)
                w.Close()
        
                tagField, _ := message.NewField("ZlibTag", tag, "")
                pack.Message.AddField(tagField)
                pack.Message.SetUuid(uuid.NewRandom())
                pack.Message.SetPayload(b.String())
                fr.Inject(pack)

                outBatch = outBatch[:0]
                f.backChan <- outBatch
        }
        wg.Done()
}

func (f *ZlibFilter) receiver(fr FilterRunner, h PluginHelper, encoder Encoder, wg *sync.WaitGroup) {
        var (
                pack *PipelinePack
                ok bool   
                e        error
        )
        ok = true
        outBatch := make([]byte, 0, 10000)
        outBytes := make([]byte, 0, 10000)
        ticker := time.Tick(time.Duration(f.FlushInterval) * time.Millisecond)
        inChan := fr.InChan()

        for ok {
                select {  
                case pack, ok = <-inChan:
                        if !ok {
                                // Closed inChan => we're shutting down, flush data
                                if len(outBatch) > 0 {
                                        f.batchChan <- outBatch
                                }
                                close(f.batchChan)
                                break
                        } 
                        f.msgLoopCount = pack.MsgLoopCount

                        if outBytes, e = encoder.Encode(pack); e != nil {
                                fr.LogError(fmt.Errorf("Error encoding message: %s", e))
                        } else {
                            if len(outBytes) > 0 {
                                outBytes = append(outBytes, '\n')
                                outBatch = append(outBatch, outBytes...)

                                if len(outBatch) > f.FlushBytes {
                                        f.batchChan <- outBatch
                                        outBatch = <-f.backChan
                                }
                            }
                            outBytes = outBytes[:0]
                        } 
                        pack.Recycle()
                case <-ticker:
                        fmt.Println("Timeout")
                        if len(outBatch) > 0 {
                        f.batchChan <- outBatch
                        outBatch = <-f.backChan
                        } 
                }
        }

        wg.Done()
}

func (f *ZlibFilter) Run(fr FilterRunner, h PluginHelper) (err error) {
        base_name := f.EncoderName
        full_name := fr.Name() + "-" + f.EncoderName
        encoder, ok := h.Encoder(base_name, full_name)
        if !ok {
            return fmt.Errorf("Encoder not found: %s", full_name)
        }

        var wg sync.WaitGroup
        wg.Add(2)
        go f.receiver(fr, h, encoder, &wg)
        go f.committer(fr, h, &wg)
        wg.Wait()

    return
}

func init() {
    RegisterPlugin("ZlibFilter", func() interface{} {
        return new(ZlibFilter)
    })
}

