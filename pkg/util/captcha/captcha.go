package captcha

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aide-family/moon/pkg/config"
	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/shapes"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/click"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"
)

type GenResult struct {
	CaptchaKey        string         `json:"captchaKey"`
	MasterImageBase64 string         `json:"masterImageBase64"`
	ThumbImageBase64  string         `json:"thumbImageBase64"`
	ThumbSize         int            `json:"thumbSize"`
	TileWidth         int            `json:"tileWidth"`
	TileHeight        int            `json:"tileHeight"`
	TileX             int            `json:"tileX"`
	TileY             int            `json:"tileY"`
	DotData           string         `json:"dotData"`
	Angle             int            `json:"angle"`
	CaptchaType       config.Captcha `json:"captchaType"`
	Expired           time.Duration  `json:"expired"`
}

func (g *GenResult) MarshalBinary() ([]byte, error) {
	return json.Marshal(g)
}

func (g *GenResult) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, g)
}

type VerifyCaptcha struct {
}

type GenerateCaptcha struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// GenerateClickCaptcha 生成点击验证码
func GenerateClickCaptcha() (*GenResult, error) {

	builder := click.NewBuilder(
		click.WithRangeLen(option.RangeVal{Min: 3, Max: 6}),
		click.WithRangeVerifyLen(option.RangeVal{Min: 2, Max: 3}),
		click.WithRangeThumbBgDistort(1),
		click.WithIsThumbNonDeformAbility(true),
	)

	shapeMaps, err := shapes.GetShapes()
	if err != nil {
		return nil, err
	}

	// background images
	imgs, err := imagesv2.GetImages()
	if err != nil {
		return nil, err
	}

	builder.SetResources(
		click.WithShapes(shapeMaps),
		click.WithBackgrounds(imgs),
	)

	shapeCapt := builder.MakeShape()

	captData, err := shapeCapt.Generate()

	if err != nil {
		return nil, err
	}

	dotData := captData.GetData()
	if dotData == nil {
		return nil, fmt.Errorf("dotData is nil")
	}

	var masterImageBase64, thumbImageBase64 string

	masterImageBase64, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}

	thumbImageBase64, err = captData.GetThumbImage().ToBase64()
	if err != nil {
		return nil, err
	}

	dotsByte, _ := json.Marshal(dotData)

	return &GenResult{
		MasterImageBase64: masterImageBase64,
		ThumbImageBase64:  thumbImageBase64,
		DotData:           string(dotsByte),
	}, nil
}

// GenerateSlideCaptcha 生成滑动验证码
func GenerateSlideCaptcha() (*GenResult, error) {

	builder := slide.NewBuilder(
	//slide.WithGenGraphNumber(2),
	//slide.WithEnableGraphVerticalRandom(true),
	)
	imgs, err := imagesv2.GetImages()
	if err != nil {

	}

	graphs, err := tiles.GetTiles()
	if err != nil {
		return nil, err
	}

	var newGraphs = make([]*slide.GraphImage, 0, len(graphs))
	for i := 0; i < len(graphs); i++ {
		graph := graphs[i]
		newGraphs = append(newGraphs, &slide.GraphImage{
			OverlayImage: graph.OverlayImage,
			MaskImage:    graph.MaskImage,
			ShadowImage:  graph.ShadowImage,
		})
	}

	// set resources
	builder.SetResources(
		slide.WithGraphImages(newGraphs),
		slide.WithBackgrounds(imgs),
	)

	slideBasicCapt := builder.Make()

	captData, err := slideBasicCapt.Generate()
	if err != nil {
		return nil, err
	}

	blockData := captData.GetData()
	if blockData == nil {
		return nil, fmt.Errorf("blockData is nil")
	}
	var masterImageBase64, tileImageBase64 string
	masterImageBase64, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}

	tileImageBase64, err = captData.GetTileImage().ToBase64()
	if err != nil {
		return nil, err
	}

	return &GenResult{
		MasterImageBase64: masterImageBase64,
		ThumbImageBase64:  tileImageBase64,
		TileWidth:         blockData.Width,
		TileHeight:        blockData.Height,
		TileX:             blockData.DX,
		TileY:             blockData.DY,
	}, nil
}

// GenerateRotateCaptcha 旋转验证码生成
func GenerateRotateCaptcha() (*GenResult, error) {

	builder := rotate.NewBuilder(
		rotate.WithRangeAnglePos([]option.RangeVal{
			{Min: 20, Max: 330},
		}),
	)

	imgs, err := imagesv2.GetImages()
	if err != nil {
		return nil, err
	}

	// set resources
	builder.SetResources(
		rotate.WithImages(imgs),
	)

	rotateBasicCapt := builder.Make()

	captData, err := rotateBasicCapt.Generate()
	if err != nil {
		return nil, err
	}

	blockData := captData.GetData()

	if blockData == nil {

		return nil, fmt.Errorf("blockData is nil")
	}

	var masterImageBase64, thumbImageBase64 string
	masterImageBase64, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}

	thumbImageBase64, err = captData.GetThumbImage().ToBase64()
	if err != nil {
		return nil, err
	}

	return &GenResult{
		MasterImageBase64: masterImageBase64,
		ThumbImageBase64:  thumbImageBase64,
		ThumbSize:         blockData.Width,
		Angle:             blockData.Angle,
	}, nil
}
