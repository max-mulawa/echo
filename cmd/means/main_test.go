package main

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	server = "localhost"
)

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func TestWritingPrices(t *testing.T) {
	for _, tc := range []struct {
		description  string
		prices       []PriceRecord
		minTime      int32
		maxTime      int32
		expectedMean int32
	}{
		{
			description: "prices with duplicated timestamp",
			prices: []PriceRecord{
				{Timestamp: 12345, Price: 101},
				{Timestamp: 12346, Price: 102},
				{Timestamp: 12347, Price: 100},
				{Timestamp: 40960, Price: 5},
			},
			minTime:      12288,
			maxTime:      16384,
			expectedMean: 101,
		},
		{
			description: "prices below zero",
			prices: []PriceRecord{
				{Timestamp: 12345, Price: -102},
				{Timestamp: 12346, Price: 102},
				{Timestamp: 12347, Price: -100},
				{Timestamp: 12348, Price: 100},
				{Timestamp: 40960, Price: 5},
			},
			minTime:      12288,
			maxTime:      16384,
			expectedMean: 0,
		},
		{
			description:  "price query with no prices records inserted",
			prices:       []PriceRecord{},
			minTime:      12288,
			maxTime:      16384,
			expectedMean: 0,
		},
		{
			description: "price total over int32 range",
			prices: []PriceRecord{
				{Timestamp: 12345, Price: math.MaxInt32 - 1},
				{Timestamp: 12346, Price: math.MaxInt32 - 1},
				{Timestamp: 12347, Price: math.MaxInt32 - 1},
				{Timestamp: 40960, Price: 5},
			},
			minTime:      12288,
			maxTime:      16384,
			expectedMean: math.MaxInt32 - 1,
		},
		{
			description: "no matching prices",
			prices: []PriceRecord{
				{Timestamp: 12345, Price: 101},
				{Timestamp: 12346, Price: 102},
				{Timestamp: 12347, Price: 100},
				{Timestamp: 40960, Price: 5},
			},
			minTime:      122880,
			maxTime:      163840,
			expectedMean: 0,
		},
		{
			description: "multi value mean",
			prices: []PriceRecord{
				{Timestamp: 570320973, Price: 9829},
				{Timestamp: 570323876, Price: 9840},
				{Timestamp: 570380248, Price: 9849},
				{Timestamp: 570430178, Price: 9863},
				{Timestamp: 570488884, Price: 9863},
				{Timestamp: 570528757, Price: 9865},
				{Timestamp: 570595235, Price: 9876},
				{Timestamp: 570609467, Price: 9890},
				{Timestamp: 570645224, Price: 9894},
				{Timestamp: 570692483, Price: 9907},
				{Timestamp: 570696617, Price: 9901},
				{Timestamp: 570743665, Price: 9905},
				{Timestamp: 570819981, Price: 9901},
				{Timestamp: 570844831, Price: 9920},
				{Timestamp: 570861054, Price: 9928},
				{Timestamp: 570949600, Price: 9931},
				{Timestamp: 571017314, Price: 9936},
				{Timestamp: 571103071, Price: 9939},
				{Timestamp: 571139451, Price: 9931},
				{Timestamp: 571166953, Price: 9943},
				{Timestamp: 571176640, Price: 9953},
				{Timestamp: 571266141, Price: 9972},
				{Timestamp: 571274584, Price: 9986},
				{Timestamp: 571295423, Price: 9996},
				{Timestamp: 571307077, Price: 10012},
				{Timestamp: 571345464, Price: 10026},
				{Timestamp: 571347949, Price: 10041},
				{Timestamp: 571399622, Price: 10047},
				{Timestamp: 571482464, Price: 10039},
				{Timestamp: 571574538, Price: 10044},
				{Timestamp: 571602464, Price: 10059},
				{Timestamp: 571627027, Price: 10069},
				{Timestamp: 571628558, Price: 10071},
				{Timestamp: 571644223, Price: 10071},
				{Timestamp: 571743994, Price: 10089},
				{Timestamp: 571784514, Price: 10107},
				{Timestamp: 571812233, Price: 10116},
				{Timestamp: 571874549, Price: 10128},
				{Timestamp: 571936719, Price: 10127},
				{Timestamp: 571991340, Price: 10141},
				{Timestamp: 572082924, Price: 10133},
				{Timestamp: 572165658, Price: 10149},
				{Timestamp: 572227220, Price: 10170},
				{Timestamp: 572319883, Price: 10172},
				{Timestamp: 572355211, Price: 10183},
				{Timestamp: 572415216, Price: 10186},
				{Timestamp: 572417685, Price: 10201},
				{Timestamp: 572421270, Price: 10217},
				{Timestamp: 572513465, Price: 10220},
				{Timestamp: 572554433, Price: 10215},
				{Timestamp: 572577221, Price: 10215},
				{Timestamp: 572639375, Price: 10217},
				{Timestamp: 572672175, Price: 10235},
				{Timestamp: 572744338, Price: 10243},
				{Timestamp: 572836506, Price: 10248},
				{Timestamp: 572871989, Price: 10255},
				{Timestamp: 572901481, Price: 10256},
				{Timestamp: 572919707, Price: 10276},
				{Timestamp: 573016491, Price: 10290},
				{Timestamp: 573034914, Price: 10282},
				{Timestamp: 573036010, Price: 10302},
				{Timestamp: 573058179, Price: 10317},
				{Timestamp: 573142986, Price: 10331},
				{Timestamp: 573173608, Price: 10330},
				{Timestamp: 573205032, Price: 10344},
				{Timestamp: 573207918, Price: 10343},
				{Timestamp: 573282443, Price: 10347},
				{Timestamp: 573291183, Price: 10368},
				{Timestamp: 573312842, Price: 10369},
				{Timestamp: 573363624, Price: 10384},
				{Timestamp: 573421286, Price: 10378},
				{Timestamp: 573479697, Price: 10386},
				{Timestamp: 573568276, Price: 10387},
				{Timestamp: 573661044, Price: 10393},
				{Timestamp: 573718923, Price: 10409},
				{Timestamp: 573774850, Price: 10411},
				{Timestamp: 573867072, Price: 10407},
				{Timestamp: 573891936, Price: 10427},
				{Timestamp: 573932749, Price: 10436},
				{Timestamp: 573989506, Price: 10454},
				{Timestamp: 574000365, Price: 10448},
				{Timestamp: 574076564, Price: 10459},
				{Timestamp: 574139759, Price: 10480},
				{Timestamp: 574156060, Price: 10477},
				{Timestamp: 574255754, Price: 10488},
				{Timestamp: 574292428, Price: 10494},
				{Timestamp: 574318040, Price: 10500},
				{Timestamp: 574406318, Price: 10493},
				{Timestamp: 574479213, Price: 10500},
				{Timestamp: 574501849, Price: 10503},
				{Timestamp: 574572996, Price: 10519},
				{Timestamp: 574671527, Price: 10536},
				{Timestamp: 574769187, Price: 10554},
				{Timestamp: 574841344, Price: 10562},
				{Timestamp: 574881320, Price: 10568},
				{Timestamp: 574937393, Price: 10565},
				{Timestamp: 574941092, Price: 10557},
				{Timestamp: 574972820, Price: 10574},
				{Timestamp: 574987170, Price: 10574},
				{Timestamp: 575046321, Price: 10576},
				{Timestamp: 575137607, Price: 10568},
				{Timestamp: 575230714, Price: 10565},
				{Timestamp: 575236241, Price: 10559},
				{Timestamp: 575274966, Price: 10578},
				{Timestamp: 575336521, Price: 10586},
				{Timestamp: 575344004, Price: 10601},
				{Timestamp: 575375086, Price: 10597},
				{Timestamp: 575420436, Price: 10610},
				{Timestamp: 575484758, Price: 10604},
				{Timestamp: 575517894, Price: 10623},
				{Timestamp: 575546594, Price: 10634},
				{Timestamp: 575636935, Price: 10630},
				{Timestamp: 575700230, Price: 10639},
				{Timestamp: 575756619, Price: 10646},
				{Timestamp: 575846537, Price: 10660},
				{Timestamp: 575857693, Price: 10657},
				{Timestamp: 575922878, Price: 10659},
				{Timestamp: 575950888, Price: 10675},
				{Timestamp: 576016081, Price: 10682},
				{Timestamp: 576039975, Price: 10678},
				{Timestamp: 576081753, Price: 10693},
				{Timestamp: 576121929, Price: 10695},
				{Timestamp: 576179856, Price: 10713},
				{Timestamp: 576278186, Price: 10717},
				{Timestamp: 576356554, Price: 10721},
				{Timestamp: 576386258, Price: 10741},
				{Timestamp: 576447926, Price: 10738},
				{Timestamp: 576502272, Price: 10745},
				{Timestamp: 576530463, Price: 10752},
				{Timestamp: 576555511, Price: 10749},
				{Timestamp: 576632727, Price: 10770},
				{Timestamp: 576718373, Price: 10769},
				{Timestamp: 576784004, Price: 10777},
				{Timestamp: 576837874, Price: 10789},
				{Timestamp: 576854972, Price: 10788},
				{Timestamp: 576947216, Price: 10786},
				{Timestamp: 576986047, Price: 10800},
				{Timestamp: 577062959, Price: 10793},
				{Timestamp: 577118686, Price: 10808},
				{Timestamp: 577215988, Price: 10806},
				{Timestamp: 577315077, Price: 10809},
				{Timestamp: 577359823, Price: 10827},
				{Timestamp: 577381027, Price: 10840},
				{Timestamp: 577445388, Price: 10851},
				{Timestamp: 577471671, Price: 10850},
				{Timestamp: 577483099, Price: 10859},
				{Timestamp: 577535260, Price: 10874},
				{Timestamp: 577583054, Price: 10870},
				{Timestamp: 577655863, Price: 10889},
				{Timestamp: 577677594, Price: 10884},
				{Timestamp: 577717517, Price: 10877},
				{Timestamp: 577746954, Price: 10872},
				{Timestamp: 577817156, Price: 10868},
				{Timestamp: 577862983, Price: 10877},
				{Timestamp: 577961382, Price: 10887},
				{Timestamp: 578039322, Price: 10903},
				{Timestamp: 578067938, Price: 10911},
				{Timestamp: 578070905, Price: 10915},
				{Timestamp: 578096877, Price: 10913},
				{Timestamp: 578190877, Price: 10905},
				{Timestamp: 578269575, Price: 10913},
				{Timestamp: 578287863, Price: 10930},
				{Timestamp: 578328674, Price: 10930},
				{Timestamp: 578349423, Price: 10938},
				{Timestamp: 578353911, Price: 10954},
				{Timestamp: 578424807, Price: 10972},
				{Timestamp: 578445592, Price: 10990},
				{Timestamp: 578468180, Price: 10992},
				{Timestamp: 578563095, Price: 10999},
				{Timestamp: 578636960, Price: 11015},
				{Timestamp: 578671116, Price: 11028},
				{Timestamp: 578704001, Price: 11049},
				{Timestamp: 578765522, Price: 11049},
				{Timestamp: 578836946, Price: 11051},
				{Timestamp: 578888932, Price: 11055},
				{Timestamp: 578925959, Price: 11058},
				{Timestamp: 578950679, Price: 11061},
				{Timestamp: 579026095, Price: 11059},
				{Timestamp: 579051643, Price: 11064},
				{Timestamp: 579114238, Price: 11075},
				{Timestamp: 579162113, Price: 11081},
				{Timestamp: 579185108, Price: 11080},
				{Timestamp: 579203141, Price: 11077},
				{Timestamp: 579233071, Price: 11093},
				{Timestamp: 579330733, Price: 11105},
				{Timestamp: 579418856, Price: 11102},
				{Timestamp: 579432932, Price: 11096},
				{Timestamp: 579513988, Price: 11094},
				{Timestamp: 579602579, Price: 11107},
				{Timestamp: 579680865, Price: 11123},
				{Timestamp: 579722316, Price: 11119},
				{Timestamp: 579733493, Price: 11112},
				{Timestamp: 579756254, Price: 11131},
				{Timestamp: 579839402, Price: 11135},
				{Timestamp: 579925380, Price: 11146},
				{Timestamp: 579981518, Price: 11153},
				{Timestamp: 580026400, Price: 11171},
				{Timestamp: 580036943, Price: 11190},
				{Timestamp: 580061660, Price: 11211},
				{Timestamp: 580119384, Price: 11230},
				{Timestamp: 580179203, Price: 11233},
				{Timestamp: 580249575, Price: 11251},
				{Timestamp: 580307185, Price: 11252},
				{Timestamp: 580319223, Price: 11251},
				{Timestamp: 580408039, Price: 11266},
				{Timestamp: 580416587, Price: 11272},
				{Timestamp: 580515253, Price: 11270},
				{Timestamp: 580545883, Price: 11285},
				{Timestamp: 580550617, Price: 11303},
				{Timestamp: 580594222, Price: 11298},
				{Timestamp: 580644407, Price: 11318},
				{Timestamp: 580717069, Price: 11323},
				{Timestamp: 580748720, Price: 11340},
				{Timestamp: 580778172, Price: 11339},
				{Timestamp: 580780318, Price: 11357},
				{Timestamp: 580826963, Price: 11355},
				{Timestamp: 580877973, Price: 11351},
				{Timestamp: 580977393, Price: 11354},
				{Timestamp: 581065897, Price: 11352},
				{Timestamp: 581071583, Price: 11367},
				{Timestamp: 581127765, Price: 11376},
				{Timestamp: 581140321, Price: 11380},
				{Timestamp: 581159090, Price: 11377},
				{Timestamp: 581235669, Price: 11373},
				{Timestamp: 581300354, Price: 11394},
				{Timestamp: 581394073, Price: 11409},
				{Timestamp: 581477266, Price: 11410},
				{Timestamp: 581562368, Price: 11418},
				{Timestamp: 581633526, Price: 11419},
			},
			minTime:      570321234,
			maxTime:      578272818,
			expectedMean: 10421,
		},
	} {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", server, serverPort))
		require.NoError(t, err)
		defer conn.Close()

		for _, p := range tc.prices {
			for _, b := range marshalPriceRecord(p) {
				conn.Write([]byte{b})
			}
		}

		conn.Write(marshalPriceQuery(PriceQuery{MinTime: tc.minTime, MaxTime: tc.maxTime}))
		meanResp := make([]byte, 4)
		conn.Read(meanResp)

		mean := UnmarshalInt32(meanResp)

		require.Equal(t, tc.expectedMean, mean)
	}

}

func marshalPriceRecord(r PriceRecord) []byte {
	record := make([]byte, 9)
	record[0] = 'I'

	MarshalInt32(r.Timestamp, record[1:5])
	MarshalInt32(r.Price, record[5:9])

	return record
}

func marshalPriceQuery(q PriceQuery) []byte {
	record := make([]byte, 9)
	record[0] = 'Q'

	MarshalInt32(q.MinTime, record[1:5])
	MarshalInt32(q.MaxTime, record[5:9])

	return record
}
