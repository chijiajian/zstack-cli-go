package common

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/query"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

func Fetch[T any](cmd *cobra.Command, params *param.QueryParam, path string, options ...query.Option) (query.Result[T], error) {
	timeout, _ := cmd.Flags().GetDuration("request-timeout")
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()
	reader, err := client.NewReadClient(ctx)
	if err != nil {
		return query.Result[T]{}, fmt.Errorf("Not logged in: %w", err)
	}
	all, _ := cmd.Flags().GetBool("all-pages")
	return query.Run(ctx, params.Values, all, func(ctx context.Context, q url.Values) (query.Result[T], error) {
		var response struct {
			Items *[]T `json:"inventories"`
			Total *int `json:"total"`
		}
		err := reader.Read(ctx, path, q, &response)
		if err != nil {
			return query.Result[T]{}, err
		}
		if (q.Get("count") == "true" || q.Get("replyWithCount") == "true") && response.Total == nil {
			return query.Result[T]{}, fmt.Errorf("server omitted query total")
		}
		if response.Total != nil && *response.Total < 0 {
			return query.Result[T]{}, fmt.Errorf("invalid negative total")
		}
		if response.Items == nil && q.Get("count") != "true" {
			return query.Result[T]{}, fmt.Errorf("server omitted query inventories")
		}
		var items []T
		if response.Items != nil {
			items = *response.Items
		}
		return query.Result[T]{Items: items, Total: response.Total}, nil
	}, options...)
}

func PrintQuery(cmd *cobra.Command, data any, total *int) error {
	output, _ := cmd.Flags().GetString("output")
	fields, _ := cmd.Flags().GetStringSlice("fields")
	format := utils.ParseFormat(output)
	count, _ := cmd.Flags().GetBool("count")
	if count {
		if total == nil {
			return fmt.Errorf("server omitted query total")
		}
		return utils.Print(*total, format)
	}
	value := reflect.ValueOf(data)
	if value.IsValid() && value.Kind() == reflect.Slice && value.IsNil() {
		data = reflect.MakeSlice(value.Type(), 0, 0).Interface()
	}
	if len(fields) > 0 {
		if err := utils.ValidateOutputFields(data, fields); err != nil {
			return err
		}
	}
	if (format == utils.JSONFormat || format == utils.YAMLFormat || format == utils.TextFormat) && len(fields) > 0 {
		var err error
		data, err = utils.ProjectRows(data, fields)
		if err != nil {
			return err
		}
		fields = nil
	}
	withTotal, _ := cmd.Flags().GetBool("reply-with-count")
	pagination, _ := cmd.Flags().GetBool("pagination")
	if (withTotal || pagination) && total != nil {
		if format == utils.JSONFormat || format == utils.YAMLFormat {
			// Project item fields before wrapping, so --fields refers to item fields.
			return utils.PrintQueryEnvelope(data, *total, format, fields)
		}
		if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Total: %d\n", *total); err != nil {
			return err
		}
	}
	return utils.PrintWithFields(data, format, fields)
}
