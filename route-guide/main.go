package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/proto"
	pb "github.com/ryutah/go-grpc-sample/route-guide/routeguide/go"
)

type routeGuideServer struct {
	savedFeatures []*pb.Feature
	mu            sync.Mutex
	routeNotes    map[string][]*pb.RouteNote
}

func newServer() *routeGuideServer {
	s := &routeGuideServer{
		routeNotes: make(map[string][]*pb.RouteNote),
	}
	s.loadFeatures("testdata/route_guide_db.json")
	return s
}

func (r *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
	for _, feature := range r.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}
	return &pb.Feature{Location: point}, nil
}

func (r *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	for _, feature := range r.savedFeatures {
		if !inRange(feature.Location, rect) {
			continue
		}
		if err := stream.Send(feature); err != nil {
			return err
		}
	}
	return nil
}

func (r *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var (
		pointCount, featureCount, distance int32
		lastPoint                          *pb.Point
	)
	startTime := time.Now()
	for {
		point, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		pointCount++
		for _, feature := range r.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}
		if lastPoint != nil {
			distance += calcDistance(lastPoint, point)
		}
		lastPoint = point
	}

	endTime := time.Now()
	return stream.SendAndClose(&pb.RouteSummary{
		PointCount:   pointCount,
		FeatureCount: featureCount,
		Distance:     distance,
		ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
	})
}

func (r *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		key := serialize(in.Location)
		r.mu.Lock()
		r.routeNotes[key] = append(r.routeNotes[key], in)
		rn := make([]*pb.RouteNote, len(r.routeNotes[key]))
		copy(rn, r.routeNotes[key])
		r.mu.Unlock()

		for _, note := range rn {
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}

func (r *routeGuideServer) loadFeatures(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to load default features: %#v", err)
	}
	if err := json.Unmarshal(file, &r.savedFeatures); err != nil {
		log.Fatalf("failed to load default features: %#v", err)
	}
}

func inRange(point *pb.Point, rect *pb.Rectangle) bool {
	var (
		left   = math.Min(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
		right  = math.Max(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
		top    = math.Max(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
		bottom = math.Min(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
	)

	return left <= float64(point.Longitude) && float64(point.Longitude) <= right &&
		bottom <= float64(point.Latitude) && float64(point.Latitude) <= top
}

func calcDistance(p1 *pb.Point, p2 *pb.Point) int32 {
	const (
		CordFactor float64 = 1e7
		R          float64 = float64(6371000) // Earth radius in meter.
	)

	var (
		lat1 = toRadians(float64(p1.Latitude) / CordFactor)
		lat2 = toRadians(float64(p2.Latitude) / CordFactor)
		lng1 = toRadians(float64(p1.Longitude) / CordFactor)
		lng2 = toRadians(float64(p2.Longitude) / CordFactor)
	)
	var (
		dlat = lat2 - lat1
		dlng = lng2 - lng1
	)

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int32(R * c)
}

func toRadians(num float64) float64 {
	return num * math.Pi / float64(180)
}

func serialize(point *pb.Point) string {
	return fmt.Sprintf("%d %d", point.Latitude, point.Longitude)
}

func main() {
	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("failed to listen: %#v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start server : %#v", err)
	}
}
