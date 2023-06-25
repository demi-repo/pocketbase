package jsvm

import (
	"encoding/json"
	"mime/multipart"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/pocketbase/pocketbase/tools/security"
)

func testBindsCount(vm *goja.Runtime, namespace string, count int, t *testing.T) {
	v, err := vm.RunString(`Object.keys(` + namespace + `).length`)
	if err != nil {
		t.Fatal(err)
	}

	total, _ := v.Export().(int64)

	if int(total) != count {
		t.Fatalf("Expected %d %s binds, got %d", count, namespace, total)
	}
}

// note: this test is useful as a reminder to update the tests in case
// a new base binding is added.
func TestBaseBindsCount(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	testBindsCount(vm, "this", 14, t)
}

func TestBaseBindsRecord(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	collection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}

	vm := goja.New()
	baseBinds(vm)
	vm.Set("collection", collection)

	// without record data
	// ---
	v1, err := vm.RunString(`new Record(collection)`)
	if err != nil {
		t.Fatal(err)
	}

	m1, ok := v1.Export().(*models.Record)
	if !ok {
		t.Fatalf("Expected m1 to be models.Record, got \n%v", m1)
	}

	// with record data
	// ---
	v2, err := vm.RunString(`new Record(collection, { email: "test@example.com" })`)
	if err != nil {
		t.Fatal(err)
	}

	m2, ok := v2.Export().(*models.Record)
	if !ok {
		t.Fatalf("Expected m2 to be models.Record, got \n%v", m2)
	}

	if m2.Collection().Name != "users" {
		t.Fatalf("Expected record with collection %q, got \n%v", "users", m2.Collection())
	}

	if m2.Email() != "test@example.com" {
		t.Fatalf("Expected record with email field set to %q, got \n%v", "test@example.com", m2)
	}
}

func TestBaseBindsCollection(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	v, err := vm.RunString(`new Collection({ name: "test", createRule: "@request.auth.id != ''", schema: [{name: "title", "type": "text"}] })`)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := v.Export().(*models.Collection)
	if !ok {
		t.Fatalf("Expected models.Collection, got %v", m)
	}

	if m.Name != "test" {
		t.Fatalf("Expected collection with name %q, got %q", "test", m.Name)
	}

	expectedRule := "@request.auth.id != ''"
	if m.CreateRule == nil || *m.CreateRule != expectedRule {
		t.Fatalf("Expected create rule %q, got %v", "@request.auth.id != ''", m.CreateRule)
	}

	if f := m.Schema.GetFieldByName("title"); f == nil {
		t.Fatalf("Expected schema to be set, got %v", m.Schema)
	}
}

func TestBaseVMAdminBind(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	v, err := vm.RunString(`new Admin({ email: "test@example.com" })`)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := v.Export().(*models.Admin)
	if !ok {
		t.Fatalf("Expected models.Admin, got %v", m)
	}
}

func TestBaseBindsSchema(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	v, err := vm.RunString(`new Schema([{name: "title", "type": "text"}])`)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := v.Export().(*schema.Schema)
	if !ok {
		t.Fatalf("Expected schema.Schema, got %v", m)
	}

	if f := m.GetFieldByName("title"); f == nil {
		t.Fatalf("Expected schema fields to be loaded, got %v", m.Fields())
	}
}

func TestBaseBindsSchemaField(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	v, err := vm.RunString(`new SchemaField({name: "title", "type": "text"})`)
	if err != nil {
		t.Fatal(err)
	}

	f, ok := v.Export().(*schema.SchemaField)
	if !ok {
		t.Fatalf("Expected schema.SchemaField, got %v", f)
	}

	if f.Name != "title" {
		t.Fatalf("Expected field %q, got %v", "title", f)
	}
}

func TestBaseBindsMailerMessage(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	v, err := vm.RunString(`new MailerMessage({
		from: {name: "test_from", address: "test_from@example.com"},
		to: [
			{name: "test_to1", address: "test_to1@example.com"},
			{name: "test_to2", address: "test_to2@example.com"},
		],
		bcc: [
			{name: "test_bcc1", address: "test_bcc1@example.com"},
			{name: "test_bcc2", address: "test_bcc2@example.com"},
		],
		cc: [
			{name: "test_cc1", address: "test_cc1@example.com"},
			{name: "test_cc2", address: "test_cc2@example.com"},
		],
		subject: "test_subject",
		html: "test_html",
		text: "test_text",
		headers: {
			header1: "a",
			header2: "b",
		}
	})`)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := v.Export().(*mailer.Message)
	if !ok {
		t.Fatalf("Expected mailer.Message, got %v", m)
	}

	raw, err := json.Marshal(m)

	expected := `{"from":{"Name":"test_from","Address":"test_from@example.com"},"to":[{"Name":"test_to1","Address":"test_to1@example.com"},{"Name":"test_to2","Address":"test_to2@example.com"}],"bcc":[{"Name":"test_bcc1","Address":"test_bcc1@example.com"},{"Name":"test_bcc2","Address":"test_bcc2@example.com"}],"cc":[{"Name":"test_cc1","Address":"test_cc1@example.com"},{"Name":"test_cc2","Address":"test_cc2@example.com"}],"subject":"test_subject","html":"test_html","text":"test_text","headers":{"header1":"a","header2":"b"},"attachments":null}`

	if string(raw) != expected {
		t.Fatalf("Expected \n%s, \ngot \n%s", expected, raw)
	}
}

func TestBaseBindsCommand(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	_, err := vm.RunString(`
		let runCalls = 0;

		let cmd = new Command({
			use: "test",
			run: (c, args) => {
				runCalls++;
			}
		});

		cmd.run(null, []);

		if (cmd.use != "test") {
			throw new Error('Expected cmd.use "test", got: ' + cmd.use);
		}

		if (runCalls != 1) {
			throw new Error('Expected runCalls 1, got: ' + runCalls);
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBaseBindsValidationError(t *testing.T) {
	vm := goja.New()
	baseBinds(vm)

	scenarios := []struct {
		js            string
		expectCode    string
		expectMessage string
	}{
		{
			`new ValidationError()`,
			"",
			"",
		},
		{
			`new ValidationError("test_code")`,
			"test_code",
			"",
		},
		{
			`new ValidationError("test_code", "test_message")`,
			"test_code",
			"test_message",
		},
	}

	for _, s := range scenarios {
		v, err := vm.RunString(s.js)
		if err != nil {
			t.Fatal(err)
		}

		m, ok := v.Export().(validation.Error)
		if !ok {
			t.Fatalf("[%s] Expected validation.Error, got %v", s.js, m)
		}

		if m.Code() != s.expectCode {
			t.Fatalf("[%s] Expected code %q, got %q", s.js, s.expectCode, m.Code())
		}

		if m.Message() != s.expectMessage {
			t.Fatalf("[%s] Expected message %q, got %q", s.js, s.expectMessage, m.Message())
		}
	}
}

func TestBaseBindsDao(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	baseBinds(vm)
	vm.Set("db", app.Dao().ConcurrentDB())
	vm.Set("db2", app.Dao().NonconcurrentDB())

	scenarios := []struct {
		js              string
		concurrentDB    dbx.Builder
		nonconcurrentDB dbx.Builder
	}{
		{
			js:              "new Dao(db)",
			concurrentDB:    app.Dao().ConcurrentDB(),
			nonconcurrentDB: app.Dao().ConcurrentDB(),
		},
		{
			js:              "new Dao(db, db2)",
			concurrentDB:    app.Dao().ConcurrentDB(),
			nonconcurrentDB: app.Dao().NonconcurrentDB(),
		},
	}

	for _, s := range scenarios {
		v, err := vm.RunString(s.js)
		if err != nil {
			t.Fatalf("[%s] Failed to execute js script, got %v", s.js, err)
		}

		d, ok := v.Export().(*daos.Dao)
		if !ok {
			t.Fatalf("[%s] Expected daos.Dao, got %v", s.js, d)
		}

		if d.ConcurrentDB() != s.concurrentDB {
			t.Fatalf("[%s] The ConcurrentDB instances doesn't match", s.js)
		}

		if d.NonconcurrentDB() != s.nonconcurrentDB {
			t.Fatalf("[%s] The NonconcurrentDB instances doesn't match", s.js)
		}
	}
}

func TestDbxBinds(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	vm.Set("db", app.Dao().DB())
	baseBinds(vm)
	dbxBinds(vm)

	testBindsCount(vm, "$dbx", 15, t)

	sceneraios := []struct {
		js       string
		expected string
	}{
		{
			`$dbx.exp("a = 1").build(db, {})`,
			"a = 1",
		},
		{
			`$dbx.hashExp({
				"a": 1,
				b: null,
				c: [1, 2, 3],
			}).build(db, {})`,
			"`a`={:p0} AND `b` IS NULL AND `c` IN ({:p1}, {:p2}, {:p3})",
		},
		{
			`$dbx.not($dbx.exp("a = 1")).build(db, {})`,
			"NOT (a = 1)",
		},
		{
			`$dbx.and($dbx.exp("a = 1"), $dbx.exp("b = 2")).build(db, {})`,
			"(a = 1) AND (b = 2)",
		},
		{
			`$dbx.or($dbx.exp("a = 1"), $dbx.exp("b = 2")).build(db, {})`,
			"(a = 1) OR (b = 2)",
		},
		{
			`$dbx.in("a", 1, 2, 3).build(db, {})`,
			"`a` IN ({:p0}, {:p1}, {:p2})",
		},
		{
			`$dbx.notIn("a", 1, 2, 3).build(db, {})`,
			"`a` NOT IN ({:p0}, {:p1}, {:p2})",
		},
		{
			`$dbx.like("a", "test1", "test2").match(true, false).build(db, {})`,
			"`a` LIKE {:p0} AND `a` LIKE {:p1}",
		},
		{
			`$dbx.orLike("a", "test1", "test2").match(false, true).build(db, {})`,
			"`a` LIKE {:p0} OR `a` LIKE {:p1}",
		},
		{
			`$dbx.notLike("a", "test1", "test2").match(true, false).build(db, {})`,
			"`a` NOT LIKE {:p0} AND `a` NOT LIKE {:p1}",
		},
		{
			`$dbx.orNotLike("a", "test1", "test2").match(false, false).build(db, {})`,
			"`a` NOT LIKE {:p0} OR `a` NOT LIKE {:p1}",
		},
		{
			`$dbx.exists($dbx.exp("a = 1")).build(db, {})`,
			"EXISTS (a = 1)",
		},
		{
			`$dbx.notExists($dbx.exp("a = 1")).build(db, {})`,
			"NOT EXISTS (a = 1)",
		},
		{
			`$dbx.between("a", 1, 2).build(db, {})`,
			"`a` BETWEEN {:p0} AND {:p1}",
		},
		{
			`$dbx.notBetween("a", 1, 2).build(db, {})`,
			"`a` NOT BETWEEN {:p0} AND {:p1}",
		},
	}

	for _, s := range sceneraios {
		result, err := vm.RunString(s.js)
		if err != nil {
			t.Fatalf("[%s] Failed to execute js script, got %v", s.js, err)
		}

		v, _ := result.Export().(string)

		if v != s.expected {
			t.Fatalf("[%s] Expected \n%s, \ngot \n%s", s.js, s.expected, v)
		}
	}
}

func TestTokensBinds(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	admin, err := app.Dao().FindAdminByEmail("test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	record, err := app.Dao().FindAuthRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	vm := goja.New()
	vm.Set("$app", app)
	vm.Set("admin", admin)
	vm.Set("record", record)
	baseBinds(vm)
	tokensBinds(vm)

	testBindsCount(vm, "$tokens", 8, t)

	sceneraios := []struct {
		js  string
		key string
	}{
		{
			`$tokens.adminAuthToken($app, admin)`,
			admin.TokenKey + app.Settings().AdminAuthToken.Secret,
		},
		{
			`$tokens.adminResetPasswordToken($app, admin)`,
			admin.TokenKey + app.Settings().AdminPasswordResetToken.Secret,
		},
		{
			`$tokens.adminFileToken($app, admin)`,
			admin.TokenKey + app.Settings().AdminFileToken.Secret,
		},
		{
			`$tokens.recordAuthToken($app, record)`,
			record.TokenKey() + app.Settings().RecordAuthToken.Secret,
		},
		{
			`$tokens.recordVerifyToken($app, record)`,
			record.TokenKey() + app.Settings().RecordVerificationToken.Secret,
		},
		{
			`$tokens.recordResetPasswordToken($app, record)`,
			record.TokenKey() + app.Settings().RecordPasswordResetToken.Secret,
		},
		{
			`$tokens.recordChangeEmailToken($app, record)`,
			record.TokenKey() + app.Settings().RecordEmailChangeToken.Secret,
		},
		{
			`$tokens.recordFileToken($app, record)`,
			record.TokenKey() + app.Settings().RecordFileToken.Secret,
		},
	}

	for _, s := range sceneraios {
		result, err := vm.RunString(s.js)
		if err != nil {
			t.Fatalf("[%s] Failed to execute js script, got %v", s.js, err)
		}

		v, _ := result.Export().(string)

		if _, err := security.ParseJWT(v, s.key); err != nil {
			t.Fatalf("[%s] Failed to parse JWT %v, got %v", s.js, v, err)
		}
	}
}

func TestSecurityRandomStringBinds(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	baseBinds(vm)
	securityBinds(vm)

	testBindsCount(vm, "$security", 7, t)

	sceneraios := []struct {
		js     string
		length int
	}{
		{`$security.randomString(6)`, 6},
		{`$security.randomStringWithAlphabet(7, "abc")`, 7},
		{`$security.pseudorandomString(8)`, 8},
		{`$security.pseudorandomStringWithAlphabet(9, "abc")`, 9},
	}

	for _, s := range sceneraios {
		result, err := vm.RunString(s.js)
		if err != nil {
			t.Fatalf("[%s] Failed to execute js script, got %v", s.js, err)
		}

		v, _ := result.Export().(string)

		if len(v) != s.length {
			t.Fatalf("[%s] Expected %d length string, \ngot \n%v", s.js, s.length, v)
		}
	}
}

func TestSecurityTokenBinds(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	baseBinds(vm)
	securityBinds(vm)

	testBindsCount(vm, "$security", 7, t)

	sceneraios := []struct {
		js       string
		expected string
	}{
		{
			`$security.parseUnverifiedToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0.aXzC7q7z1lX_hxk5P0R368xEU7H1xRwnBQQcLAmG0EY")`,
			`{"name":"John Doe","sub":"1234567890"}`,
		},
		{
			`$security.parseToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0.aXzC7q7z1lX_hxk5P0R368xEU7H1xRwnBQQcLAmG0EY", "test")`,
			`{"name":"John Doe","sub":"1234567890"}`,
		},
		{
			`$security.createToken({"exp": 123}, "test", 0)`, // overwrite the exp claim for static token
			`"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEyM30.7gbv7w672gApdBRASI6OniCtKwkKjhieSxsr6vxSrtw"`,
		},
	}

	for _, s := range sceneraios {
		result, err := vm.RunString(s.js)
		if err != nil {
			t.Fatalf("[%s] Failed to execute js script, got %v", s.js, err)
		}

		raw, _ := json.Marshal(result.Export())

		if string(raw) != s.expected {
			t.Fatalf("[%s] Expected \n%s, \ngot \n%s", s.js, s.expected, raw)
		}
	}
}

func TestFilesystemBinds(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	vm.Set("mh", &multipart.FileHeader{Filename: "test"})
	vm.Set("testFile", filepath.Join(app.DataDir(), "data.db"))
	baseBinds(vm)
	filesystemBinds(vm)

	testBindsCount(vm, "$filesystem", 3, t)

	// fileFromPath
	{
		v, err := vm.RunString(`$filesystem.fileFromPath(testFile)`)
		if err != nil {
			t.Fatal(err)
		}

		file, _ := v.Export().(*filesystem.File)

		if file == nil || file.OriginalName != "data.db" {
			t.Fatalf("[fileFromPath] Expected file with name %q, got %v", file.OriginalName, file)
		}
	}

	// fileFromBytes
	{
		v, err := vm.RunString(`$filesystem.fileFromBytes([1, 2, 3], "test")`)
		if err != nil {
			t.Fatal(err)
		}

		file, _ := v.Export().(*filesystem.File)

		if file == nil || file.OriginalName != "test" {
			t.Fatalf("[fileFromBytes] Expected file with name %q, got %v", file.OriginalName, file)
		}
	}

	// fileFromMultipart
	{
		v, err := vm.RunString(`$filesystem.fileFromMultipart(mh)`)
		if err != nil {
			t.Fatal(err)
		}

		file, _ := v.Export().(*filesystem.File)

		if file == nil || file.OriginalName != "test" {
			t.Fatalf("[fileFromMultipart] Expected file with name %q, got %v", file.OriginalName, file)
		}
	}
}

func TestFormsBinds(t *testing.T) {
	vm := goja.New()
	formsBinds(vm)

	testBindsCount(vm, "this", 20, t)
}

func TestApisBindsCount(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	apisBinds(vm)

	testBindsCount(vm, "this", 7, t)
	testBindsCount(vm, "$apis", 10, t)
}

func TestApisBindsRoute(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	baseBinds(vm)
	apisBinds(vm)

	_, err := vm.RunString(`
		let handlerCalls = 0;

		const route = new Route({
			method: "test_method",
			path: "test_path",
			handler: () => {
				handlerCalls++;
			}
		});

		route.handler();

		if (handlerCalls != 1) {
			throw new Error('Expected handlerCalls 1, got ' + handlerCalls);
		}

		if (route.method != "test_method") {
			throw new Error('Expected method "test_method", got ' + route.method);
		}

		if (route.path != "test_path") {
			throw new Error('Expected method "test_path", got ' + route.path);
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestApisBindsApiError(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	apisBinds(vm)

	scenarios := []struct {
		js            string
		expectCode    int
		expectMessage string
		expectData    string
	}{
		{"new ApiError()", 0, "", "null"},
		{"new ApiError(100, 'test', {'test': 1})", 100, "Test.", `{"test":1}`},
		{"new NotFoundError()", 404, "The requested resource wasn't found.", "null"},
		{"new NotFoundError('test', {'test': 1})", 404, "Test.", `{"test":1}`},
		{"new BadRequestError()", 400, "Something went wrong while processing your request.", "null"},
		{"new BadRequestError('test', {'test': 1})", 400, "Test.", `{"test":1}`},
		{"new ForbiddenError()", 403, "You are not allowed to perform this request.", "null"},
		{"new ForbiddenError('test', {'test': 1})", 403, "Test.", `{"test":1}`},
		{"new UnauthorizedError()", 401, "Missing or invalid authentication token.", "null"},
		{"new UnauthorizedError('test', {'test': 1})", 401, "Test.", `{"test":1}`},
	}

	for _, s := range scenarios {
		v, err := vm.RunString(s.js)
		if err != nil {
			t.Errorf("[%s] %v", s.js, err)
			continue
		}

		apiErr, ok := v.Export().(*apis.ApiError)
		if !ok {
			t.Errorf("[%s] Expected ApiError, got %v", s.js, v)
			continue
		}

		if apiErr.Code != s.expectCode {
			t.Errorf("[%s] Expected Code %d, got %d", s.js, s.expectCode, apiErr.Code)
		}

		if apiErr.Message != s.expectMessage {
			t.Errorf("[%s] Expected Message %q, got %q", s.js, s.expectMessage, apiErr.Message)
		}

		dataRaw, _ := json.Marshal(apiErr.RawData())
		if string(dataRaw) != s.expectData {
			t.Errorf("[%s] Expected Data %q, got %q", s.js, s.expectData, dataRaw)
		}
	}
}

func TestLoadingDynamicModel(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	baseBinds(vm)
	dbxBinds(vm)
	vm.Set("$app", app)

	_, err := vm.RunString(`
		let result = new DynamicModel({
			text:        "",
			bool:        false,
			number:      0,
			select_many: [],
			json:        [],
			// custom map-like field
			obj: {},
		})

		$app.dao().db()
		    .select("text", "bool", "number", "select_many", "json", "('{\"test\": 1}') as obj")
		    .from("demo1")
		    .where($dbx.hashExp({"id": "84nmscqy84lsi1t"}))
		    .limit(1)
		    .one(result)

		if (result.text != "test") {
			throw new Error('Expected text "test", got ' + result.text);
		}

		if (result.bool != true) {
			throw new Error('Expected bool true, got ' + result.bool);
		}

		if (result.number != 123456) {
			throw new Error('Expected number 123456, got ' + result.number);
		}

		if (result.select_many.length != 2 || result.select_many[0] != "optionB" || result.select_many[1] != "optionC") {
			throw new Error('Expected select_many ["optionB", "optionC"], got ' + result.select_many);
		}

		if (result.json.length != 3 || result.json[0] != 1 || result.json[1] != 2 || result.json[2] != 3) {
			throw new Error('Expected json [1, 2, 3], got ' + result.json);
		}

		if (result.obj.get("test") != 1) {
			throw new Error('Expected obj.get("test") 1, got ' + JSON.stringify(result.obj));
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadingArrayOf(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	vm := goja.New()
	baseBinds(vm)
	dbxBinds(vm)
	vm.Set("$app", app)

	_, err := vm.RunString(`
		let result = $arrayOf(new DynamicModel({
			id:   "",
			text: "",
		}))

		$app.dao().db()
		    .select("id", "text")
		    .from("demo1")
		    .where($dbx.exp("id='84nmscqy84lsi1t' OR id='al1h9ijdeojtsjy'"))
		    .limit(2)
		    .orderBy("text ASC")
		    .all(result)

		if (result.length != 2) {
			throw new Error('Expected 2 list items, got ' + result.length);
		}

		if (result[0].id != "84nmscqy84lsi1t") {
			throw new Error('Expected 0.id "84nmscqy84lsi1t", got ' + result[0].id);
		}
		if (result[0].text != "test") {
			throw new Error('Expected 0.text "test", got ' + result[0].text);
		}

		if (result[1].id != "al1h9ijdeojtsjy") {
			throw new Error('Expected 1.id "al1h9ijdeojtsjy", got ' + result[1].id);
		}
		if (result[1].text != "test2") {
			throw new Error('Expected 1.text "test2", got ' + result[1].text);
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
}
