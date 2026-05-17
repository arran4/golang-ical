import sys

content = sys.stdin.read()

content = content.replace('''                  <dtstart>
            <parameters>
              <value>
                <text>DATE</text>
              </value>
            </parameters>
            <date>2008-10-06</date>
          </dtstart>''', '''                  <dtstart>
            <date>2008-10-06</date>
          </dtstart>''')

with open("calendar_xml_test.go", "w") as f:
    f.write(content)
