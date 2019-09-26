Account.all.each do |account|
  puts "select #{account.id} (#{account.name})"
  ActiveRecord::Base.connection.schema_search_path = "#{account.schema}, hstore"
  ActiveRecord::Base.connection.tables.each do |table|
    next if table.match(/\Aschema_migrations\Z/)     
    puts "#{table}"
    ActiveRecord::Base.connection.execute("DELETE FROM #{table};")
  end
end
