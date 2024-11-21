package mockclickhouse

type Row struct {
	scanCallback func(dest ...any) error
}

func (r *Row) SetScan(cb func(dest ...any) error) {
	r.scanCallback = cb
}

func (r *Row) Err() error {
	return nil
}

func (r *Row) Scan(dest ...any) error {
	if r.scanCallback != nil {
		err := r.scanCallback(dest...)
		return err
	}
	return nil
}

func (r *Row) ScanStruct(dest any) error {
	return nil
}
