package templates

var Resolver = `package gen

import (
	"context"
	
	"github.com/novacloudcz/graphql-orm/resolvers"
	uuid "github.com/satori/go.uuid"
)

type Resolver struct {
	DB *DB
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}
{{range .Model.Objects}}
func (r *Resolver) {{.Name}}ResultType() {{.Name}}ResultTypeResolver {
	return &{{.LowerName}}ResultTypeResolver{r}
}
{{if .HasRelationships}}
func (r *Resolver) {{.Name}}() {{.Name}}Resolver {
	return &{{.LowerName}}Resolver{r}
}
{{end}}
{{end}}

type mutationResolver struct{ *Resolver }

{{range .Model.Objects}}
func (r *mutationResolver) Create{{.Name}}(ctx context.Context, input  map[string]interface{}) (item *{{.Name}}, err error) {
	item = &{{.Name}}{ID:uuid.Must(uuid.NewV4()).String()}
	tx := r.DB.db.Begin()
{{range $rel := .Relationships}}
{{if $rel.IsToMany}}
	if ids,ok:=input["{{$rel.Name}}Ids"].([]interface{}); ok {
		items := []{{$rel.TargetType}}{}
		tx.Find(&items, "id IN (?)", ids)
		association := tx.Model(&item).Association("{{$rel.MethodName}}")
		association.Replace(items)
	}
{{end}}
{{end}}

	err = resolvers.CreateItem(ctx, tx, item, input)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit().Error
	return 
}
func (r *mutationResolver) Update{{.Name}}(ctx context.Context, id string, input  map[string]interface{}) (item *{{.Name}}, err error) {
	item = &{{.Name}}{}
	tx := r.DB.db.Begin()
	
	err = resolvers.GetItem(ctx, tx, item, &id)
	if err != nil {
		return 
	}

{{range $rel := .Relationships}}
{{if $rel.IsToMany}}
	if ids,ok:=input["{{$rel.Name}}Ids"].([]interface{}); ok {
		items := []{{$rel.TargetType}}{}
		tx.Find(&items, "id IN (?)", ids)
		association := tx.Model(&item).Association("{{$rel.MethodName}}")
		association.Replace(items)
	}
{{end}}
{{end}}

	err = resolvers.UpdateItem(ctx, tx, item, input)

	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit().Error
	return 
}
func (r *mutationResolver) Delete{{.Name}}(ctx context.Context, id string) (item *{{.Name}}, err error) {
	item = &{{.Name}}{}
	err = resolvers.GetItem(ctx, r.DB.Query(), item, &id)
	if err != nil {
		return 
	}

	err = resolvers.DeleteItem(ctx, r.DB.Query(), item, id)

	return 
}
{{end}}

type queryResolver struct{ *Resolver }

{{range $object := .Model.Objects}}
func (r *queryResolver) {{$object.Name}}(ctx context.Context, id *string, q *string) (*{{$object.Name}}, error) {
	t := {{$object.Name}}{}
	err := resolvers.GetItem(ctx, r.DB.Query(), &t, id)
	return &t, err
}
func (r *queryResolver) {{$object.PluralName}}(ctx context.Context, offset *int, limit *int, q *string) (*{{$object.Name}}ResultType, error) {
	return &{{$object.Name}}ResultType{}, nil
}

type {{$object.LowerName}}ResultTypeResolver struct{ *Resolver }

func (r *{{$object.LowerName}}ResultTypeResolver) Items(ctx context.Context, obj *{{$object.Name}}ResultType) (items []*{{$object.Name}}, err error) {
	err = resolvers.GetResultTypeItems(ctx, r.DB.db, &items)
	return
}

func (r *{{$object.LowerName}}ResultTypeResolver) Count(ctx context.Context, obj *{{$object.Name}}ResultType) (count int, err error) {
	return resolvers.GetResultTypeCount(ctx, r.DB.db, &{{$object.Name}}{})
}

{{if .HasRelationships}}
type {{$object.LowerName}}Resolver struct { *Resolver }

{{range $index, $relationship := .Relationships}}
func (r *{{$object.LowerName}}Resolver) {{$relationship.MethodName}}(ctx context.Context, obj *{{$object.Name}}) (res {{.ReturnType}}, err error) {
{{if $relationship.IsToMany}}
	items := []*{{.TargetType}}{}
	err = r.DB.Query().Model(obj).Related(&items, "{{$relationship.MethodName}}").Error
	res = items
{{else}}
	item := {{.TargetType}}{}
	_res := r.DB.Query().Model(obj).Related(&item, "{{$relationship.MethodName}}")
	if _res.RecordNotFound() {
		return
	} else {
		err = _res.Error
	}
	res = &item
{{end}}
	return 
}
{{end}}
{{end}}

{{end}}
`
