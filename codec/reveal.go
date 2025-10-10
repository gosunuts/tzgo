// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/trilitech/tzgo/tezos"
)

// Reveal represents "reveal" operation
type Reveal struct {
	Manager
	PublicKey tezos.Key `json:"public_key"`
	Proof     tezos.Signature
}

func (o Reveal) Kind() tezos.OpType {
	return tezos.OpTypeReveal
}

func (o Reveal) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"public_key":`)
	buf.WriteString(strconv.Quote(o.PublicKey.String()))
	if o.PublicKey.Type == tezos.KeyTypeBls12_381 && o.Proof.Data != nil {
		buf.WriteString(`,"proof":`)
		buf.WriteString(strconv.Quote(o.Proof.String()))
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o Reveal) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	buf.Write(o.PublicKey.Bytes())

	if p.Version >= tezos.Versions[tezos.PtSeouLo] {
		if o.Proof.Type == tezos.SignatureTypeBls12_381 {
			buf.WriteByte(0x01)
			buf.Write(o.Proof.Bytes())
		} else {
			buf.WriteByte(0x00)
		}
	}

	return nil
}

func (o *Reveal) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	if err = o.PublicKey.DecodeBuffer(buf); err != nil {
		return
	}

	if p.Version >= tezos.Versions[tezos.PtSeouLo] {
		b, _ := buf.ReadByte()
		if o.PublicKey.Type == tezos.KeyTypeBls12_381 {
			if b != 0x01 {
				err = fmt.Errorf("tz4 reveal requires proof flag | %d", b)
				return
			}
			var sig tezos.Signature
			if err := sig.DecodeBuffer(buf); err != nil {
				return err
			}
			o.Proof = sig
		} else if b != 0x00 {
			err = fmt.Errorf("tz1/2/3 reveal must not contain proof | %d", b)
			return
		}
	}
	return
}

func (o Reveal) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *Reveal) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
