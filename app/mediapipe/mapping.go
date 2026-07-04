package main

import (
	"encoding/json"
	"marmalade/internal/errs"
	"os"

	"github.com/vladimirvivien/go4vl/v4l2"
)

type Mapping struct {
	Format  string
	FourCC  v4l2.FourCCType
	CodecID uint32
	PixFmt  int32
}

func find_mapping(pixfmt v4l2.FourCCType) (*Mapping, error) {
	file, err := os.Open("fourcc.json")
	if err != nil {
		return nil, errs.CreateError("opening FourCC file", err)
	}

	var mapping []Mapping

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&mapping)
	if err != nil {
		return nil, errs.CreateError("reading FourCC file", err)
	}

	for _, pixfmt_map := range mapping {
		if pixfmt_map.FourCC == pixfmt {
			return &pixfmt_map, nil
		}
	}

	return nil, nil
}
